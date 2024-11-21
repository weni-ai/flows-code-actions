package rabbitmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnection(t *testing.T) {
	url := "amqp://127.0.0.1:5672/"

	conn := NewEDA(url)
	assert.NotNil(t, conn)
	assert.NotNil(t, conn.conn)
	ch, err := conn.conn.Channel(context.TODO())
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	consumer := &Consumer{
		ExchangeName: "test_exchange",
		QueueName:    "test_queue",
		Handler:      &ConsumerHandler{},
	}

	conn.AddConsumer(consumer)

	err = conn.StartConsumers()
	assert.NoError(t, err)
}

type ConsumerHandler struct {
}

func (h *ConsumerHandler) Handle(ctx context.Context, msg []byte) error {
	// Handle the message
	return nil
}
