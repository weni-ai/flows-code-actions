package project

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/eventdriven"
	"github.com/weni-ai/flows-code-actions/internal/eventdriven/rabbitmq"
	"github.com/weni-ai/flows-code-actions/internal/permission"
)

const (
	EXCHANGE_NAME = "projects.topic"
	QUEUE_NAME    = "code-actions.projects"
)

type ProjectConsumer struct {
	rabbitmq.Consumer
	projectService    UseCase
	permissionService permission.UserPermissionUseCase
}

func NewProjectConsumer(projectService UseCase, permissionService permission.UserPermissionUseCase, exchange, queue string) *ProjectConsumer {
	exchangeName := EXCHANGE_NAME
	if exchange != "" {
		exchangeName = exchange
	}
	queueName := QUEUE_NAME
	if queue != "" {
		queueName = queue
	}
	c := &ProjectConsumer{
		Consumer: rabbitmq.Consumer{
			QueueName:    queueName,
			ExchangeName: exchangeName,
		},
		projectService:    projectService,
		permissionService: permissionService,
	}
	c.Handler = c
	return c
}

func (c *ProjectConsumer) Handle(ctx context.Context, eventMsg []byte) error {
	var evt eventdriven.ProjectEvent
	err := json.Unmarshal(eventMsg, &evt)
	if err != nil {
		log.Errorf("Error unmarshalling event: %v", err)
		return err
	}

	newProject := NewProject(evt.UUID, evt.Name)
	if _, err := c.projectService.Create(ctx, newProject); err != nil {
		if err.Error() != "project already exists" {
			return errors.Wrapf(err, "Error creating project on handle event by EDA consumer for project: %v", newProject)
		}
		log.Error(errors.Wrapf(err, "Error creating project on handle event by EDA consumer for project: %v", newProject))
	}
	for _, auths := range evt.Authorizations {
		userPerm := permission.NewUserPermission(evt.UUID, auths.UserEmail, permission.Role(auths.Role))
		if _, err := c.permissionService.Create(ctx, userPerm); err != nil {
			return errors.Wrapf(err, "Error creating user permission on handle event by EDA consumer for user: %v", userPerm)
		}
	}

	log.Infof("project consumer handled event: %v", evt)
	return nil
}
