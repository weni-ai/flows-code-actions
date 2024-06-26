package code

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewCodeService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, code *Code) (*Code, error) {
	return s.repo.Create(ctx, code)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Code, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListProjectCodes(ctx context.Context, projectUUID string, codeType string) ([]Code, error) {
	return s.repo.ListByProjectUUID(ctx, projectUUID, codeType)
}

func (s *Service) Update(ctx context.Context, id string, name string, source string, codeType string) (*Code, error) {
	code, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		code.Name = name
	}
	if source != "" {
		code.Source = source
	}
	if codeType != "" {
		t := CodeType(codeType)
		if err := t.Validate(); err != nil {
			return nil, err
		}
		code, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		code.Type = t
	}
	return s.repo.Update(ctx, id, code)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
