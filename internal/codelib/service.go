package codelib

import "context"

type Service struct {
	repo Repository
}

func NewCodeLibService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, codelib *CodeLib) (*CodeLib, error) {
	return s.repo.Create(ctx, codelib)
}

func (s *Service) CreateBulk(ctx context.Context, codelibs []*CodeLib) ([]*CodeLib, error) {
	return s.repo.CreateBulk(ctx, codelibs)
}

func (s *Service) List(ctx context.Context, language *LanguageType) ([]CodeLib, error) {
	return s.repo.List(ctx, language)
}

func (s *Service) Find(ctx context.Context, name string, language *LanguageType) (*CodeLib, error) {
	return s.repo.Find(ctx, name, language)
}
