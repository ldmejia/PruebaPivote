package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ldmejia/PruebaPivote/internal/domain/account"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

const accountColumns = "id, number, owner_id, currency, balance, blocked, created_at"

func scanAccount(row interface{ Scan(...any) error }) (*account.Account, error) {
	var a account.Account
	err := row.Scan(&a.ID, &a.Number, &a.OwnerID, &a.Currency, &a.Balance, &a.Blocked, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, account.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AccountRepository) Create(ctx context.Context, a *account.Account) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `INSERT INTO accounts (number, owner_id, currency, balance)
OUTPUT INSERTED.id, INSERTED.created_at
VALUES (@p1, @p2, @p3, @p4)`, a.Number, a.OwnerID, a.Currency, a.Balance).Scan(&a.ID, &a.CreatedAt)
	if isDuplicateKey(err) {
		return account.ErrDuplicatedNumber
	}
	if err != nil {
		return err
	}
	if a.Balance > 0 {
		if err := insertMovement(ctx, tx, a.Number, "", transfer.MovementOpening, a.Balance, a.Balance); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *AccountRepository) GetByNumber(ctx context.Context, number string) (*account.Account, error) {
	return scanAccount(r.db.QueryRowContext(ctx, "SELECT "+accountColumns+" FROM accounts WHERE number = @p1", number))
}

func (r *AccountRepository) ListByOwner(ctx context.Context, ownerID string) ([]*account.Account, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+accountColumns+" FROM accounts WHERE owner_id = @p1 ORDER BY created_at", ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*account.Account{}
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *AccountRepository) UpdateBlocked(ctx context.Context, a *account.Account) error {
	res, err := r.db.ExecContext(ctx, "UPDATE accounts SET blocked = @p1, updated_at = SYSUTCDATETIME() WHERE number = @p2", a.Blocked, a.Number)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return account.ErrAccountNotFound
	}
	return nil
}
