package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/ldmejia/PruebaPivote/internal/domain/audit"
	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

type AuditStore struct {
	coll *mongo.Collection
}

func NewAuditStore(db *mongo.Database) *AuditStore {
	return &AuditStore{coll: db.Collection("audit_events")}
}

func (s *AuditStore) Save(ctx context.Context, ev transfer.Event, routingKey string) error {
	_, err := s.coll.InsertOne(ctx, bson.M{
		"_id":         ev.ID,
		"type":        ev.Type,
		"routing_key": routingKey,
		"occurred_at": ev.OccurredAt,
		"data":        ev.Data,
		"received_at": time.Now().UTC(),
	})
	if mongo.IsDuplicateKeyError(err) {
		return audit.ErrDuplicatedEvent
	}
	return err
}
