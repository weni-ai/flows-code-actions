package project

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/furdarius/rabbitroutine"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestConsumer(t *testing.T) {

	url := "amqp://127.0.0.1:5672/"

	conn, err := amqp.Dial(url)
	assert.NoError(t, err)

	ch, err := conn.Channel()
	assert.NoError(t, err)

	defer ch.Close()
	defer ch.QueueDelete(QUEUE_NAME, false, false, false)
	defer ch.ExchangeDelete(EXCHANGE_NAME, false, false)

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

	// Bind the queue to the exchange
	err = ch.QueueBind(
		q.Name,
		"#",
		EXCHANGE_NAME,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue to exchange: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := `{
		"uuid": "f13b48e7-fbe9-4e68-975f-01ab4b09d877",
		"name": "asadf",
		"is_template": false,
		"user_email": "rafael.soares@weni.ai",
		"date_format": "D",
		"template_type_uuid": null,
		"timezone": "America/Sao_Paulo",
		"organization_id": 420209,
		"extra_fields": {},
		"authorizations": [
			{
				"user_email": "rafael.soares@weni.ai",
				"role": 3
			}
		],
		"description": "none",
		"organization_uuid": "865aa22d-4c75-4d40-bcd8-9468180f7f63",
		"brain_on": true
	}`

	err = ch.PublishWithContext(ctx,
		EXCHANGE_NAME,
		"", // send message to exchange to be consumed by any queue binded to it
		true,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(body), //TODO: handle
		},
	)
	assert.NoError(t, err)

	// Finish publishing

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

	projectService := NewProjectService(NewMemProjectRepository())

	consumer := NewProjectConsumer(projectService)

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

	proj, _ := projectService.FindByUUID(context.TODO(), "f13b48e7-fbe9-4e68-975f-01ab4b09d877")

	assert.NotNil(t, proj)
	assert.Equal(t, "f13b48e7-fbe9-4e68-975f-01ab4b09d877", proj.UUID)
}

type inMemoryRepo struct {
	projects map[string]*Project
}

func NewMemProjectRepository() *inMemoryRepo {
	return &inMemoryRepo{
		projects: make(map[string]*Project),
	}
}

func (r *inMemoryRepo) Create(ctx context.Context, p *Project) (*Project, error) {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	p.ID = primitive.NewObjectID()
	r.projects[p.UUID] = p
	return p, nil
}

func (r *inMemoryRepo) FindByUUID(ctx context.Context, uuid string) (*Project, error) {
	return r.projects[uuid], nil
}

func (r *inMemoryRepo) Update(ctx context.Context, p *Project) (*Project, error) {
	for _, pr := range r.projects {
		if pr.ID == p.ID {
			p.UpdatedAt = time.Now()
			r.projects[p.UUID] = p
			return p, nil
		}
	}
	return nil, errors.New("error project not found")
}
