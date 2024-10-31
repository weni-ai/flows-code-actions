package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/weni-ai/flows-code-actions/internal/eventdriven"
	"github.com/weni-ai/flows-code-actions/internal/eventdriven/rabbitmq"
)

const (
	EXCHANGE_NAME = "permissions.topic"
	QUEUE_NAME    = "code-actions-permissions"
)

type PermissionConsumer struct {
	rabbitmq.Consumer
	permissionService UserPermissionUseCase
}

func NewPermissionConsumer(permissionService UserPermissionUseCase) *PermissionConsumer {
	c := &PermissionConsumer{
		Consumer: rabbitmq.Consumer{
			QueueName:    QUEUE_NAME,
			ExchangeName: EXCHANGE_NAME,
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
		log.Printf("Error unmarshalling event: %v", err)
		return err
	}

	userPerm := NewUserPermission(evt.Project, evt.User, Role(evt.Role))

	switch evt.Action {
	case "create":
		if _, err := c.permissionService.Create(ctx, userPerm); err != nil {
			return err
		}
	case "update":
		finded, err := c.permissionService.Find(ctx, userPerm)
		if err != nil {
			return err
		}
		if finded == nil || finded.ProjectUUID != evt.Project {
			if _, err := c.permissionService.Create(ctx, userPerm); err != nil {
				return err
			}
		} else {
			if _, err := c.permissionService.Update(ctx, finded.ID.Hex(), userPerm); err != nil {
				return err
			}
		}
	case "delete":
		finded, err := c.permissionService.Find(ctx, userPerm)
		if err != nil {
			return err
		}
		if finded.ProjectUUID == evt.Project {
			if err := c.permissionService.Delete(ctx, finded.ID.Hex()); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("invalid action: %s", evt.Action)
	}
	return nil
}
