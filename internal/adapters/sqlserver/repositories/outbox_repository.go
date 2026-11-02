package repository

import (
	"context"
	"database/sql"

	"github.com/ldmejia/PruebaPivote/internal/adapters/rabbitmq"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) PublishPending(ctx context.Context, limit int, publish func(ctx context.Context, msg rabbitmq.OutboxMessage) error) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	list, err := lockPending(ctx, tx, limit)
	if err != nil {
		return 0, err
	}

	published := 0
	var publishErr error
	for _, p := range list {
		if publishErr = publish(ctx, p.msg); publishErr != nil {
			break
		}
		if _, err := tx.ExecContext(ctx, "UPDATE outbox SET published_at = SYSUTCDATETIME() WHERE id = @p1", p.id); err != nil {
			return 0, err
		}
		published++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return published, publishErr
}

type pendingEvent struct {
	id  int64
	msg rabbitmq.OutboxMessage
}

func lockPending(ctx context.Context, tx *sql.Tx, limit int) ([]pendingEvent, error) {
	rows, err := tx.QueryContext(ctx, `SELECT TOP (@p1) id, LOWER(CONVERT(VARCHAR(36), event_id)), event_type, payload
FROM outbox WITH (UPDLOCK, READPAST, ROWLOCK)
WHERE published_at IS NULL
ORDER BY id`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []pendingEvent
	for rows.Next() {
		var p pendingEvent
		var payload string
		if err := rows.Scan(&p.id, &p.msg.EventID, &p.msg.Type, &payload); err != nil {
			return nil, err
		}
		p.msg.Payload = []byte(payload)
		list = append(list, p)
	}
	return list, rows.Err()
}
