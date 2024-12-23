package coderun

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewCodeRunService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, code *CodeRun) (*CodeRun, error) {
	return s.repo.Create(ctx, code)
}

func (s *Service) GetByID(ctx context.Context, id string) (*CodeRun, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByCodeID(ctx context.Context, codeID string, filter map[string]interface{}) ([]CodeRun, error) {
	return s.repo.ListByCodeID(ctx, codeID, filter)
}

func (s *Service) Update(ctx context.Context, id string, codeRun *CodeRun) (*CodeRun, error) {
	return s.repo.Update(ctx, id, codeRun)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
