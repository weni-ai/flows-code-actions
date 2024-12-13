package code

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
	"go.mongodb.org/mongo-driver/mongo"
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

	// TODO: move this lib management to another place
	if code.Language == TypePy {
		externalLibs := codelib.ExtractPythonLibs(code.Source)
		if len(externalLibs) > 0 {
			err := codelib.InstallPythonLibs(externalLibs)
			if err != nil {
				return nil, err
			}
		}

		// TODO: refactor this asap, is much responsibility to this function
		for _, lib := range externalLibs {
			codeLang := string(code.Language)
			libLang := codelib.LanguageType(codeLang)
			libFound, err := s.libService.Find(ctx, lib, &libLang)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
			if libFound == nil {
				newLib := codelib.NewCodeLib(lib, libLang)
				_, err := s.libService.Create(ctx, newLib)
				if err != nil {
					return nil, err
				}
			}
		}
	}

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
