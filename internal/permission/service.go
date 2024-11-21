package permission

import "context"

type UserService struct {
	repo UserPermissionRepository
}

func NewUserService(repo UserPermissionRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, user *UserPermission) (*UserPermission, error) {
	return s.repo.Create(ctx, user)
}

func (s *UserService) Find(ctx context.Context, user *UserPermission) (*UserPermission, error) {
	return s.repo.Find(ctx, user)
}

func (s *UserService) Update(ctx context.Context, userID string, user *UserPermission) (*UserPermission, error) {
	return s.repo.Update(ctx, userID, user)
}

func (s *UserService) Delete(ctx context.Context, userID string) error {
	return s.repo.Delete(ctx, userID)
}
