package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

const (
	Exchange     = "transfers"
	DeadExchange = "transfers.dlx"
	AuditQueue   = "audit.transfers"
	AuditDLQ     = "audit.transfers.dlq"
)

func DeclareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(Exchange, "topic", true, false, false, false, nil)
}

func DeclareAuditQueues(ch *amqp.Channel) error {
	if err := DeclareExchange(ch); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(DeadExchange, "fanout", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(AuditDLQ, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(AuditDLQ, "", DeadExchange, false, nil); err != nil {
		return err
	}
	args := amqp.Table{"x-dead-letter-exchange": DeadExchange}
	if _, err := ch.QueueDeclare(AuditQueue, true, false, false, false, args); err != nil {
		return err
	}
	return ch.QueueBind(AuditQueue, "transfer.*", Exchange, false, nil)
}
