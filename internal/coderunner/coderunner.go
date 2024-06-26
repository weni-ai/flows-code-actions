package coderunner

import (
	"context"

	"github.com/weni-ai/flows-code-actions/internal/coderun"
)

type UseCase interface {
	RunCode(ctx context.Context, codeID string, code string, language string) (*coderun.CodeRun, error)
}
