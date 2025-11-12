package event

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	conn     *amqp.Connection
	exchange string
}

func NewRabbitPublisher(dsn, exchange string) (*RabbitPublisher, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &RabbitPublisher{
		conn:     conn,
		exchange: exchange,
	}, nil
}

func (p *RabbitPublisher) Publish(ctx context.Context, evt Event) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("rabbitmq channel: %w", err)
	}
	defer ch.Close()

	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return ch.PublishWithContext(ctx, p.exchange, "", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	})
}

func (p *RabbitPublisher) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

type RabbitConsumer struct {
	conn     *amqp.Connection
	exchange string
	queue    string
}

func NewRabbitConsumer(dsn, exchange, queue string) (*RabbitConsumer, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(queue, "", exchange, false, nil); err != nil {
		return nil, fmt.Errorf("queue bind: %w", err)
	}

	return &RabbitConsumer{
		conn:     conn,
		exchange: exchange,
		queue:    queue,
	}, nil
}

func (c *RabbitConsumer) Consume(ctx context.Context, handler Handler) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("rabbitmq channel: %w", err)
	}
	defer ch.Close()

	deliveries, err := ch.Consume(c.queue, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("queue consume: %w", err)
	}

	for {
		select {
		case d, ok := <-deliveries:
			if !ok {
				return nil
			}
			var evt Event
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				continue
			}
			if handler != nil {
				if err := handler(ctx, evt); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *RabbitConsumer) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
