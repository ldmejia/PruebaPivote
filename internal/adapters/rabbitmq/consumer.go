package rabbitmq

import (
	"context"
	"errors"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	auditapp "github.com/ldmejia/PruebaPivote/internal/application/audit"
)

type Consumer struct {
	url string
	uc  *auditapp.RecordEventUseCase
}

func NewConsumer(url string, uc *auditapp.RecordEventUseCase) *Consumer {
	return &Consumer{url: url, uc: uc}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		err := c.consume(ctx)
		if ctx.Err() != nil {
			return
		}
		log.Printf("se perdio la conexion con rabbitmq, reintentando: %v", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (c *Consumer) consume(ctx context.Context) error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := DeclareAuditQueues(ch); err != nil {
		return err
	}
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	msgs, err := ch.Consume(AuditQueue, "auditor", false, false, false, false, nil)
	if err != nil {
		return err
	}
	closed := conn.NotifyClose(make(chan *amqp.Error, 1))
	log.Printf("auditor consumiendo %s", AuditQueue)
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-closed:
			return err
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("el canal de mensajes se cerro")
			}
			c.handle(msg)
		}
	}
}

// si no se pudo guardar, el nack sin requeue lo manda a la dlq y no frena a los demas
func (c *Consumer) handle(msg amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.uc.Execute(ctx, msg.Body, msg.RoutingKey); err != nil {
		log.Printf("no se pudo registrar el evento, se manda a la dlq: %v", err)
		msg.Nack(false, false)
		return
	}
	msg.Ack(false)
}
