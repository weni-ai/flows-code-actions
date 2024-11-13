package permission

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type inMemoryRepo struct {
	permissions map[string]*UserPermission
}

func NewMemPermissionRepository() *inMemoryRepo {
	return &inMemoryRepo{permissions: make(map[string]*UserPermission)}
}

func (r *inMemoryRepo) Create(ctx context.Context, p *UserPermission) (*UserPermission, error) {
	found, ok := r.permissions[p.Email]
	if ok && found.ProjectUUID == p.ProjectUUID {
		return nil, errors.New("user permission already exists")
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	p.ID = primitive.NewObjectID()
	r.permissions[p.Email] = p
	return p, nil
}

func (r *inMemoryRepo) Find(ctx context.Context, p *UserPermission) (*UserPermission, error) {
	return r.permissions[p.Email], nil
}

func (r *inMemoryRepo) Update(ctx context.Context, id string, p *UserPermission) (*UserPermission, error) {
	r.permissions[p.Email] = p
	return nil, nil
}

func (r *inMemoryRepo) Delete(ctx context.Context, userID string) error {
	delete(r.permissions, userID)
	return nil
}
