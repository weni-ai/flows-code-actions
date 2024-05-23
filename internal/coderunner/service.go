package coderunner

import (
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Service struct {
}

func NewCodeRunnerService() *Service {
	return &Service{}
}

func (s *Service) RunCode(ctx context.Context, code string, language string) (string, error) {
	var out []byte
	var err error
	switch language {
	case "python":
		out, err = exec.Command("python", "-c", code).Output()
	case "javascript":
		out, err = exec.Command("node", "-c", code).Output()
	case "go":
		out, err = exec.Command("python", "-c", code).Output()
	default:
		return "", errors.New("unsupported language code type")
	}
	if err != nil {
		return "", errors.Wrap(err, "error on executing code")
	}

	result := strings.TrimSpace(string(out))
	return result, nil
}
