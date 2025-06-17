package code

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
)

const maxSourecBytes = 1024 * 1024

type Service struct {
	repo       Repository
	libService codelib.UseCase
	conf       *config.Config
}

func NewCodeService(conf *config.Config, repo Repository, libService codelib.UseCase) *Service {
	return &Service{conf: conf, repo: repo, libService: libService}
}

func (s *Service) Create(ctx context.Context, code *Code) (*Code, error) {
	if len(code.Source) >= maxSourecBytes {
		return nil, errors.New("source code is too big")
	}

	blacklist := s.conf.GetBlackListTerms()
	for _, term := range blacklist {
		if strings.Contains(code.Source, term) {
			return nil, errors.New("source code contains blacklisted term")
		}
	}

	code.SetTimeout(code.Timeout)

	return s.repo.Create(ctx, code)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Code, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListProjectCodes(ctx context.Context, projectUUID string, codeType string) ([]Code, error) {
	return s.repo.ListByProjectUUID(ctx, projectUUID, codeType)
}

func (s *Service) Update(ctx context.Context, id string, name string, source string, codeType string, timeout int) (*Code, error) {
	if len(source) >= maxSourecBytes {
		return nil, errors.New("source code is too big")
	}

	blacklist := s.conf.GetBlackListTerms()
	for _, term := range blacklist {
		if strings.Contains(source, term) {
			return nil, errors.New("source code contains blacklisted term")
		}
	}

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
	if timeout > 0 {
		code.SetTimeout(timeout)
	}

	return s.repo.Update(ctx, id, code)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func removeItens(from []string, source []string) []string {
	mp := make(map[string]bool)
	for _, item := range source {
		mp[item] = true
	}

	newSL := []string{}
	for _, item := range from {
		if !mp[item] {
			newSL = append(newSL, item)
		}
	}

	return newSL
}
