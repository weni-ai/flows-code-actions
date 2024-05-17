package codelog

import "context"

type Service struct {
	repo Repository
}

func NewCodeLogService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, codelog *CodeLog) (*CodeLog, error) {
	return s.repo.Create(ctx, codelog)
}

func (s *Service) GetByID(ctx context.Context, id string) (*CodeLog, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListRunLogs(ctx context.Context, id string) ([]CodeLog, error) {
	return s.repo.ListRunLogs(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, content string) (*CodeLog, error) {
	return s.repo.Update(ctx, id, content)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
