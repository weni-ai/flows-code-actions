package permission

import (
	"context"
)

type UserPermissionService struct {
	repo UserPermissionRepository
}

func NewUserPermissionService(repo UserPermissionRepository) *UserPermissionService {
	return &UserPermissionService{repo: repo}
}

func (s *UserPermissionService) Create(ctx context.Context, user *UserPermission) (*UserPermission, error) {
	return s.repo.Create(ctx, user)
}

func (s *UserPermissionService) Find(ctx context.Context, user *UserPermission) (*UserPermission, error) {
	return s.repo.Find(ctx, user)
}

func (s *UserPermissionService) Update(ctx context.Context, userID string, user *UserPermission) (*UserPermission, error) {
	return s.repo.Update(ctx, userID, user)
}

func (s *UserPermissionService) Delete(ctx context.Context, userID string) error {
	return s.repo.Delete(ctx, userID)
}
