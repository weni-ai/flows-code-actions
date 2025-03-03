package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

var ErrInvalidMsg = errors.New("invalid msg")

type IConsumerHandler interface {
	Handle(context.Context, []byte) error
}

type Consumer struct {
	ExchangeName string
	QueueName    string
	Handler      IConsumerHandler
}

func (c *Consumer) Declare(ctx context.Context, ch *amqp.Channel) error {
	err := ch.ExchangeDeclarePassive(
		c.ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.WithError(err).Error("failed to declare exchange")
		return err
	}

	_, err = ch.QueueDeclarePassive(
		c.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.WithError(err).Error("failed to declare queue")
		return err
	}

	err = ch.QueueBind(
		c.QueueName,
		"#",
		c.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		log.WithError(err).Error("failed to bind queue to exchange")
		return err
	}

	return nil
}

func (c *Consumer) Consume(ctx context.Context, ch *amqp.Channel) error {
	err := ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		log.WithError(err).Error("failed to set qos")
		return err
	}

	msgs, err := ch.Consume(
		c.QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("failed to consume %v", c.QueueName))
	}

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return amqp.ErrClosed
			}

			// Handle consume
			{
				err := c.Handler.Handle(ctx, msg.Body)
				if err != nil {
					log.WithError(err).Error(fmt.Sprintf("failed to handle %v", c.QueueName))
					if !errors.Is(err, ErrInvalidMsg) {
						return err
					}
				}
			}

			err := msg.Ack(false)
			if err != nil {
				log.WithError(err).Error("failed to ack message")
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
