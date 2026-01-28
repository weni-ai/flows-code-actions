package secrets

import (
	"context"

	"github.com/pkg/errors"
)

type Service struct {
	repo Repository
}

func NewSecretService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, secret *Secret) (*Secret, error) {
	if secret.Name == "" {
		return nil, errors.New("secret name is required")
	}
	if secret.Value == "" {
		return nil, errors.New("secret value is required")
	}
	if secret.CodeID == "" {
		return nil, errors.New("code_id is required")
	}

	return s.repo.Create(ctx, secret)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Secret, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByCodeID(ctx context.Context, codeID string) ([]Secret, error) {
	return s.repo.GetByCodeID(ctx, codeID)
}

func (s *Service) Update(ctx context.Context, id string, name string, value string) (*Secret, error) {
	secret, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		secret.Name = name
	}
	if value != "" {
		secret.Value = value
	}

	return s.repo.Update(ctx, id, secret)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
