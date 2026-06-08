package mq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"goshop/internal/config"
)

const (
	ExchangeOrders   = "goshop.orders"
	QueueOrderPaid   = "goshop.order.paid"
	RoutingOrderPaid = "order.paid"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

type OrderPaidEvent struct {
	TenantID uint64 `json:"tenant_id"`
	OrderNo  string `json:"order_no"`
	PayNo    string `json:"pay_no"`
	Channel  string `json:"channel"`
	Amount   int64  `json:"amount"`
}

func NewPublisher(cfg *config.RabbitMQConfig) (*Publisher, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(ExchangeOrders, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	if _, err := ch.QueueDeclare(QueueOrderPaid, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	if err := ch.QueueBind(QueueOrderPaid, RoutingOrderPaid, ExchangeOrders, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) PublishOrderPaid(ctx context.Context, evt OrderPaidEvent) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(ctx, ExchangeOrders, RoutingOrderPaid, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
}

func (p *Publisher) Close() error {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewConsumer(cfg *config.RabbitMQConfig) (*Consumer, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Consumer{conn: conn, ch: ch}, nil
}

func (c *Consumer) ConsumeOrderPaid(handler func(OrderPaidEvent) error) error {
	if err := c.ch.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := c.ch.Consume(QueueOrderPaid, "goshop-server", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			var evt OrderPaidEvent
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				_ = d.Nack(false, false)
				continue
			}
			if err := handler(evt); err != nil {
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}
	}()
	return nil
}

func (c *Consumer) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
