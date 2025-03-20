package secret

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, secret *Secret) (*Secret, error) {
	return s.repo.Create(ctx, secret)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Secret, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByProjectUUID(ctx context.Context, projectUUID string) ([]Secret, error) {
	return s.repo.ListByProjectUUID(ctx, projectUUID)
}

func (s *Service) Update(ctx context.Context, id string, secret *Secret) (*Secret, error) {
	return s.repo.Update(ctx, id, secret)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
