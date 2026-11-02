package audit

import (
	"context"
	"errors"

	"github.com/ldmejia/PruebaPivote/internal/domain/transfer"
)

var (
	ErrDuplicatedEvent = errors.New("evento ya registrado")
	ErrInvalidEvent    = errors.New("evento invalido")
)

type Store interface {
	Save(ctx context.Context, ev transfer.Event, routingKey string) error
}
