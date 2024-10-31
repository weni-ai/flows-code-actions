package rabbitmq

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/furdarius/rabbitroutine"
	"github.com/sirupsen/logrus"
)

type EDA struct {
	conn      *rabbitroutine.Connector
	consumers []rabbitroutine.Consumer
}

func NewEDA(url string) *EDA {
	conn := rabbitroutine.NewConnector(rabbitroutine.Config{Wait: 2 * time.Second})

	conn.AddRetriedListener(func(r rabbitroutine.Retried) {
		log.Printf("try to connect to RabbitMQ: attempt=%d, error=\"%v\"",
			r.ReconnectAttempt, r.Error)
	})

	conn.AddDialedListener(func(_ rabbitroutine.Dialed) {
		log.Printf("RabbitMQ connection successfully established")
	})

	conn.AddAMQPNotifiedListener(func(n rabbitroutine.AMQPNotified) {
		log.Printf("RabbitMQ error received: %v", n.Error)
	})

	go func() {
		err := conn.Dial(context.Background(), url)
		if err != nil {
			logrus.Error("failed to establish RabbitMQ connection")
		}
	}()

	return &EDA{
		conn: conn,
	}
}

func (c *EDA) AddConsumer(consumer rabbitroutine.Consumer) {
	c.consumers = append(c.consumers, consumer)
}

func (c *EDA) StartConsumers() error {
	if c.conn == nil {
		return errors.New("EDA not initialized")
	}
	if len(c.consumers) == 0 {
		return errors.New("no consumers added")
	}
	for _, consumer := range c.consumers {
		cons := consumer
		go func() {
			log.Printf("starting consumer for queue: %s", cons)
			err := c.conn.StartConsumer(context.Background(), cons)
			if err != nil {
				logrus.Error(err)
			}
		}()
	}
	return nil
}
