package project

import (
	"context"
	"encoding/json"
	"log"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/eventdriven"
	"github.com/weni-ai/flows-code-actions/internal/eventdriven/rabbitmq"
)

const (
	EXCHANGE_NAME = "projects.topic"
	QUEUE_NAME    = "code-actions.projects"
)

type ProjectConsumer struct {
	rabbitmq.Consumer
	projectService UseCase
}

func NewProjectConsumer(projectService UseCase) *ProjectConsumer {
	c := &ProjectConsumer{
		Consumer: rabbitmq.Consumer{
			QueueName:    QUEUE_NAME,
			ExchangeName: EXCHANGE_NAME,
		},
		projectService: projectService,
	}
	c.Handler = c
	return c
}

func (c *ProjectConsumer) Handle(ctx context.Context, eventMsg []byte) error {
	var evt eventdriven.ProjectEvent
	err := json.Unmarshal(eventMsg, &evt)
	if err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return err
	}

	newProject := NewProject(evt.UUID, evt.Name)
	if _, err := c.projectService.Create(ctx, newProject); err != nil {
		return errors.Wrap(err, "Error creating project on handle event by EDA consumer")
	}
	// Manage Role from Authorization

	return nil
}
