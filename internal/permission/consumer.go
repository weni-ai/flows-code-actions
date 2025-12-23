package permission

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/internal/eventdriven"
	"github.com/weni-ai/flows-code-actions/internal/eventdriven/rabbitmq"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	EXCHANGE_NAME = "permissions.topic"
	QUEUE_NAME    = "code-actions-permissions"
)

type PermissionConsumer struct {
	rabbitmq.Consumer
	permissionService UserPermissionUseCase
}

func NewPermissionConsumer(permissionService UserPermissionUseCase, exchange, queue string) *PermissionConsumer {
	exchangeName := EXCHANGE_NAME
	if exchange != "" {
		exchangeName = exchange
	}
	queueName := QUEUE_NAME
	if queue != "" {
		queueName = queue
	}
	c := &PermissionConsumer{
		Consumer: rabbitmq.Consumer{
			QueueName:    queueName,
			ExchangeName: exchangeName,
		},
		permissionService: permissionService,
	}
	c.Handler = c
	return c
}

func (c *PermissionConsumer) Handle(ctx context.Context, eventMsg []byte) error {
	var evt eventdriven.PermissionEvent
	err := json.Unmarshal(eventMsg, &evt)
	if err != nil {
		log.Errorf("Error unmarshalling event: %v", err)
		return err
	}

	userPerm := NewUserPermission(evt.Project, evt.User, Role(evt.Role))

	switch evt.Action {
	case "create":
		if _, err := c.permissionService.Create(ctx, userPerm); err != nil {
			if err.Error() != "user permission already exists" {
				return err
			}
		}
	case "update":
		finded, err := c.permissionService.Find(ctx, userPerm)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
	if finded == nil || finded.ProjectUUID != evt.Project {
		if _, err := c.permissionService.Create(ctx, userPerm); err != nil {
			return err
		}
	} else {
		if _, err := c.permissionService.Update(ctx, finded.ID, userPerm); err != nil {
			return err
		}
	}
	case "delete":
		finded, err := c.permissionService.Find(ctx, userPerm)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil
			}
			return err
		}
	if finded.ProjectUUID == evt.Project {
		if err := c.permissionService.Delete(ctx, finded.ID); err != nil {
			return err
		}
	}
		return nil
	default:
		return errors.Wrapf(rabbitmq.ErrInvalidMsg, "action: %s, for event: %s", evt.Action, eventMsg)
	}
	log.Infof("permission consumer handled event: %v", evt)

	return nil
}
