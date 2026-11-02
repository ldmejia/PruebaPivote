package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

var errNegativeBalance = errors.New("el saldo quedaria negativo")

type TransferRepository struct {
	db *sql.DB
}

func NewTransferRepository(db *sql.DB) *TransferRepository {
	return &TransferRepository{db: db}
}

func (r *TransferRepository) FindByIdempotencyKey(ctx context.Context, ownerID, key string) (*transfer.Transfer, string, error) {
	var t transfer.Transfer
	var hash string
	var reason sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT LOWER(CONVERT(VARCHAR(36), t.id)), t.from_account, t.to_account, t.amount, t.currency, t.status, t.reason, t.created_at, k.request_hash
FROM idempotency_keys k
JOIN transfers t ON t.id = k.transfer_id
WHERE k.owner_id = @p1 AND k.idem_key = @p2`, vc(ownerID), vc(key)).Scan(&t.ID, &t.FromAccount, &t.ToAccount, &t.Amount, &t.Currency, &t.Status, &reason, &t.CreatedAt, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", transfer.ErrTransferNotFound
	}
	if err != nil {
		return nil, "", err
	}
	t.Reason = reason.String
	return &t, hash, nil
}

func (r *TransferRepository) Execute(ctx context.Context, o transfer.Order) (*transfer.Transfer, error) {
	var t *transfer.Transfer
	var err error
	for i := 0; i < 3; i++ {
		t, err = r.execute(ctx, o)
		if !isDeadlock(err) {
			return t, err
		}
	}
	return t, err
}

func (r *TransferRepository) execute(ctx context.Context, o transfer.Order) (*transfer.Transfer, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	from, to, err := lockAccounts(ctx, tx, o.FromAccount, o.ToAccount)
	if err != nil {
		return nil, err
	}
	t := &transfer.Transfer{
		ID:          o.ID,
		FromAccount: o.FromAccount,
		ToAccount:   o.ToAccount,
		Amount:      o.Amount,
		Currency:    from.Currency,
		Status:      transfer.StatusCompleted,
		CreatedAt:   o.CreatedAt,
	}
	day := shared.LocalDay(o.CreatedAt)
	used, err := dailyUsage(ctx, tx, from.Number, day)
	if err != nil {
		return nil, err
	}
	reason := transfer.Check(from, to, o.Amount, used)
	if reason == "" {
		reason = o.CoreReason
	}
	if reason != "" {
		t.Status, t.Reason = transfer.StatusRejected, reason
	}
	if err := insertTransfer(ctx, tx, t); err != nil {
		return nil, err
	}
	if t.Status == transfer.StatusCompleted {
		if err := moveMoney(ctx, tx, t); err != nil {
			return nil, err
		}
		if err := addDailyUsage(ctx, tx, from.Number, day, o.Amount); err != nil {
			return nil, err
		}
	}
	if err := insertOutbox(ctx, tx, transfer.NewEvent(t)); err != nil {
		return nil, err
	}
	if err := saveIdempotencyKey(ctx, tx, o, t.ID); err != nil {
		return nil, err
	}
	return t, tx.Commit()
}

// las cuentas se bloquean siempre en el mismo orden para no tener deadlocks
func lockAccounts(ctx context.Context, tx *sql.Tx, fromNumber, toNumber string) (*account.Account, *account.Account, error) {
	first, second := fromNumber, toNumber
	if second < first {
		first, second = second, first
	}
	a, err := lockAccount(ctx, tx, first)
	if err != nil {
		return nil, nil, err
	}
	b, err := lockAccount(ctx, tx, second)
	if err != nil {
		return nil, nil, err
	}
	if a.Number == fromNumber {
		return a, b, nil
	}
	return b, a, nil
}

func lockAccount(ctx context.Context, tx *sql.Tx, number string) (*account.Account, error) {
	return scanAccount(tx.QueryRowContext(ctx, "SELECT "+accountColumns+" FROM accounts WITH (UPDLOCK, ROWLOCK) WHERE number = @p1", vc(number)))
}

func insertTransfer(ctx context.Context, tx *sql.Tx, t *transfer.Transfer) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO transfers (id, from_account, to_account, amount, currency, status, reason, created_at)
VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8)`, t.ID, vc(t.FromAccount), vc(t.ToAccount), t.Amount, vc(t.Currency), vc(t.Status), nullString(t.Reason), t.CreatedAt)
	return err
}

func moveMoney(ctx context.Context, tx *sql.Tx, t *transfer.Transfer) error {
	fromBalance, err := addBalance(ctx, tx, t.FromAccount, -t.Amount)
	if err != nil {
		return err
	}
	toBalance, err := addBalance(ctx, tx, t.ToAccount, t.Amount)
	if err != nil {
		return err
	}
	if err := insertMovement(ctx, tx, t.FromAccount, t.ID, account.MovementDebit, t.Amount, fromBalance); err != nil {
		return err
	}
	return insertMovement(ctx, tx, t.ToAccount, t.ID, account.MovementCredit, t.Amount, toBalance)
}

func addBalance(ctx context.Context, tx *sql.Tx, number string, delta int64) (int64, error) {
	var balance int64
	err := tx.QueryRowContext(ctx, `UPDATE accounts SET balance = balance + @p1, updated_at = SYSUTCDATETIME()
OUTPUT INSERTED.balance
WHERE number = @p2 AND balance + @p1 >= 0`, delta, vc(number)).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errNegativeBalance
	}
	return balance, err
}

func insertMovement(ctx context.Context, tx *sql.Tx, number, transferID, movementType string, amount, balanceAfter int64) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO movements (account_number, transfer_id, type, amount, balance_after)
VALUES (@p1, @p2, @p3, @p4, @p5)`, vc(number), nullString(transferID), vc(movementType), amount, balanceAfter)
	return err
}

func dailyUsage(ctx context.Context, tx *sql.Tx, number string, day time.Time) (int64, error) {
	var total int64
	err := tx.QueryRowContext(ctx, `SELECT total FROM daily_usage WITH (UPDLOCK, HOLDLOCK)
WHERE account_number = @p1 AND usage_date = @p2`, vc(number), day.Format(time.DateOnly)).Scan(&total)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return total, err
}

func addDailyUsage(ctx context.Context, tx *sql.Tx, number string, day time.Time, amount int64) error {
	_, err := tx.ExecContext(ctx, `UPDATE daily_usage SET total = total + @p3 WHERE account_number = @p1 AND usage_date = @p2;
IF @@ROWCOUNT = 0
    INSERT INTO daily_usage (account_number, usage_date, total) VALUES (@p1, @p2, @p3);`, vc(number), day.Format(time.DateOnly), amount)
	return err
}

func insertOutbox(ctx context.Context, tx *sql.Tx, ev transfer.Event) error {
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO outbox (event_id, event_type, payload) VALUES (@p1, @p2, @p3)", ev.ID, vc(ev.Type), string(payload))
	return err
}

func saveIdempotencyKey(ctx context.Context, tx *sql.Tx, o transfer.Order, transferID string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO idempotency_keys (owner_id, idem_key, request_hash, transfer_id)
VALUES (@p1, @p2, @p3, @p4)`, vc(o.OwnerID), vc(o.IdempotencyKey), vc(o.RequestHash), transferID)
	if isDuplicateKey(err) {
		return transfer.ErrDuplicatedKey
	}
	return err
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
