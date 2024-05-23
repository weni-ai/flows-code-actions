package coderunner

import "context"

type UseCase interface {
	RunCode(ctx context.Context, code string, language string) (string, error)
}
