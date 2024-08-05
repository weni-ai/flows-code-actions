package code

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

const maxSourecBytes = 1024 * 1024

type Service struct {
	repo Repository
}

func NewCodeService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, code *Code) (*Code, error) {
	if len(code.Source) >= maxSourecBytes {
		return nil, errors.New("source code is too big")
	}

	if code.Language == TypePy {
		externalLibs := extractPythonLibs(code.Source)
		if len(externalLibs) > 0 {
			err := installPythonLibs(externalLibs)
			if err != nil {
				return nil, err
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

func extractPythonLibs(pythonCode string) []string {
	standardLibraries := []string{"base64", "datetime", "json", "math", "os", "random", "re", "sys"}
	re := regexp.MustCompile(`^(from|import)\s+([\w.]+)`)

	var libraries []string
	for _, line := range strings.Split(pythonCode, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 2 {
			library := strings.Split(matches[2], ".")[0]

			if !contains(standardLibraries, library) {
				libraries = append(libraries, library)
			}
		}
	}

	return removeDoubles(libraries)
}

func installPythonLibs(libs []string) error {
	for _, lib := range libs {
		cmd := exec.Command("pip", "install", lib)
		err := cmd.Run()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error on install lib: %s", lib))
		}
	}
	return nil
}

func contains(s []string, e string) bool {
	i := sort.SearchStrings(s, e)
	return i < len(s) && s[i] == e
}

func removeDoubles(s []string) []string {
	sort.Strings(s)
	j := 0
	for i := 1; i < len(s); i++ {
		if s[j] != s[i] {
			j++
			s[j] = s[i]
		}
	}
	if len(s) > 0 {
		return s[:j+1]
	}
	return s
}
