package rabbitmq

import (
	"context"
	"errors"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OutboxMessage struct {
	EventID string
	Type    string
	Payload []byte
}

type OutboxStore interface {
	PublishPending(ctx context.Context, limit int, publish func(ctx context.Context, msg OutboxMessage) error) (int, error)
}

type Relay struct {
	store    OutboxStore
	url      string
	interval time.Duration
	conn     *amqp.Connection
	ch       *amqp.Channel
}

func NewRelay(store OutboxStore, url string, interval time.Duration) *Relay {
	return &Relay{store: store, url: url, interval: interval}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	defer r.close()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

func (r *Relay) tick(ctx context.Context) {
	if err := r.connect(); err != nil {
		log.Printf("rabbitmq no disponible, los eventos quedan en el outbox: %v", err)
		return
	}
	n, err := r.store.PublishPending(ctx, 100, r.publish)
	if n > 0 {
		log.Printf("eventos publicdos: %d", n)
	}
	if err != nil {
		log.Printf("error publicando el outbox: %v", err)
		r.close()
	}
}

func (r *Relay) publish(ctx context.Context, msg OutboxMessage) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	confirm, err := r.ch.PublishWithDeferredConfirmWithContext(ctx, Exchange, msg.Type, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    msg.EventID,
		Timestamp:    time.Now(),
		Body:         msg.Payload,
	})
	if err != nil {
		return err
	}
	ok, err := confirm.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("rabbitmq rechazo el mensaje")
	}
	return nil
}

func (r *Relay) connect() error {
	if r.conn != nil && !r.conn.IsClosed() && r.ch != nil && !r.ch.IsClosed() {
		return nil
	}
	r.close()
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	if err := ch.Confirm(false); err != nil {
		conn.Close()
		return err
	}
	if err := DeclareExchange(ch); err != nil {
		conn.Close()
		return err
	}
	r.conn, r.ch = conn, ch
	return nil
}

func (r *Relay) close() {
	if r.ch != nil {
		r.ch.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
	r.conn, r.ch = nil, nil
}
