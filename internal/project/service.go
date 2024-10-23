package project

import (
	"context"
)

type Service struct {
	repo    Repository
	service UseCase
}

func NewProjectService(repo Repository, service UseCase) *Service {
	return &Service{repo: repo, service: service}
}

func (s *Service) Create(ctx context.Context, project *Project) (*Project, error) {
	return s.repo.Create(ctx, project)
}

func (s *Service) FindByUUID(ctx context.Context, project *Project) (*Project, error) {
	return s.service.FindByUUID(ctx, project)
}

func (s *Service) Update(ctx context.Context, project *Project) (*Project, error) {
	return s.repo.Update(ctx, project)
}
