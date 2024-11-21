package project

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewProjectService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, project *Project) (*Project, error) {
	return s.repo.Create(ctx, project)
}

func (s *Service) FindByUUID(ctx context.Context, uuid string) (*Project, error) {
	return s.repo.FindByUUID(ctx, uuid)
}

func (s *Service) Update(ctx context.Context, project *Project) (*Project, error) {
	return s.repo.Update(ctx, project)
}
