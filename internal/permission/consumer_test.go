package permission

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/furdarius/rabbitroutine"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestConsumer(t *testing.T) {
	url := "amqp://127.0.0.1:5672/"

	conn, err := amqp.Dial(url)
	assert.NoError(t, err)

	ch, err := conn.Channel()
	assert.NoError(t, err)

	defer ch.QueueDelete(QUEUE_NAME, false, false, false)
	defer ch.ExchangeDelete(EXCHANGE_NAME, false, false)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		EXCHANGE_NAME,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	assert.NoError(t, err)

	q, err := ch.QueueDeclare(
		QUEUE_NAME,
		true,
		false,
		false,
		false,
		nil,
	)
	assert.NoError(t, err)

	err = ch.QueueBind(
		q.Name,
		"#",
		EXCHANGE_NAME,
		false,
		nil,
	)
	assert.NoError(t, err)

	event := `{
		"action": "update",
		"project": "fa147fa6-5af0-4d99-9c00-043c89d97392",
		"user": "rafael.soares@weni.ai",
		"role": 2
	}
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		EXCHANGE_NAME,
		"#",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(event),
		},
	)
	assert.NoError(t, err)

	cconn := rabbitroutine.NewConnector(rabbitroutine.Config{Wait: 2 * time.Second})

	cconn.AddRetriedListener(func(r rabbitroutine.Retried) {
		log.Printf("try to connect to RabbitMQ: attempt=%d, error=\"%v\"",
			r.ReconnectAttempt, r.Error)
	})

	cconn.AddDialedListener(func(_ rabbitroutine.Dialed) {
		log.Printf("RabbitMQ connection successfully established")
	})

	cconn.AddAMQPNotifiedListener(func(n rabbitroutine.AMQPNotified) {
		log.Printf("RabbitMQ error received: %v", n.Error)
	})

	permissionService := NewUserPermissionService(NewMemPermissionRepository())
	consumer := NewPermissionConsumer(permissionService, "", "")

	conctx := context.Background()

	go func() {
		log.Println("conn.Start starting")
		defer log.Println("conn.Start finished")

		cconn.Dial(conctx, url)
	}()

	go func() {
		log.Println("consumer.Consume starting")
		defer log.Println("consumer.Consume finished")

		cconn.StartConsumer(conctx, consumer)
	}()

	time.Sleep(1 * time.Second)

	perm, _ := permissionService.Find(context.TODO(), &UserPermission{
		Email: "rafael.soares@weni.ai",
	})

	assert.NotNil(t, perm)
	assert.Equal(t, "rafael.soares@weni.ai", perm.Email)
}
