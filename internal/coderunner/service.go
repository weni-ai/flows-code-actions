package coderunner

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
)

type Service struct {
	codeRun *coderun.Service
}

func NewCodeRunnerService(coderun *coderun.Service) *Service {
	return &Service{codeRun: coderun}
}

func (s *Service) RunCode(ctx context.Context, codeID string, code string, language string) (*coderun.CodeRun, error) {
	newCodeRun, err := s.codeRun.Create(ctx, coderun.NewCodeRun(codeID, coderun.StatusStarted))
	if err != nil {
		return nil, err
	}

	var result string

	switch language {
	case "python":
		result, err = runPython(ctx, code)
	case "javascript":
		result, err = runJs(ctx, code)
	case "go":
		result, err = runGo(ctx, code)
	default:
		return nil, errors.New("unsupported language code type")
	}
	if err != nil {
		newCodeRun.Status = coderun.StatusFailed
		newCodeRun.Result = errors.Wrap(err, "error on executing code").Error()
		return s.codeRun.Update(ctx, codeID, newCodeRun)
	}

	newCodeRun.Result = result
	newCodeRun.Status = coderun.StatusCompleted
	return s.codeRun.Update(ctx, codeID, newCodeRun)
}

func runPython(ctx context.Context, code string) (string, error) {
	cmd := exec.Command("python", "-c", code)
	codeBuffer := bytes.NewBufferString(code)
	cmd.Stdin = codeBuffer

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("process took too long. out: %s, err: %s", stdout.String(), stderr.String())
		}
	}
	if stderr.String() != "" {
		return "", fmt.Errorf("error executing code: %s", stderr.String())
	}
	return stdout.String(), nil
}

func runJs(ctx context.Context, code string) (string, error) {
	cmd := exec.Command("node", "-e", code)
	codeBuffer := bytes.NewBufferString(code)
	cmd.Stdin = codeBuffer

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("process took too long. out: %s, err: %s", stdout.String(), stderr.String())
		}
	}
	if stderr.String() != "" {
		return "", fmt.Errorf("error executing code: %s", stderr.String())
	}
	return stdout.String(), nil
}

func runGo(ctx context.Context, code string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "coderunner")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(tmpDir+"/main.go", []byte(code), 0644); err != nil {
		return "", fmt.Errorf("error creating temp file %q: %v", tmpDir, err)
	}

	goCache := filepath.Join(tmpDir, "gocache")
	cmd := exec.CommandContext(ctx, "go", "run", tmpDir+"/main.go")
	var goPath string
	cmd.Env = []string{"GOCACHE=" + goCache}

	useModules := false
	if useModules {
		// Create a GOPATH just for modules to be downloaded
		// into GOPATH/pkg/mod.
		goPath, err = os.MkdirTemp("", "gopath")
		if err != nil {
			return "", fmt.Errorf("error creating temp directory: %v", err)
		}
		defer os.RemoveAll(goPath)
		cmd.Env = append(cmd.Env, "GO111MODULE=on", "GOPROXY=https://proxy.golang.org")
	} else {
		goPath = os.Getenv("GOPATH")
		cmd.Env = append(cmd.Env, "GO111MODULE=off")
	}

	cmd.Env = append(cmd.Env, "GOPATH="+goPath)
	var recOut bytes.Buffer
	var recErr bytes.Buffer
	cmd.Stdout = &recOut
	cmd.Stderr = &recErr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("process took too long. out: %s, err: %s", recOut.String(), recErr.String())
		}
	}
	if recErr.String() != "" {
		return "", fmt.Errorf("error executing code: %s", recErr.String())
	}
	return recOut.String(), nil
}

// ------------------------------------------------

func compileGoAndRun(code string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "coderunner")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	// var buildPkgArg = "."

	if err := ioutil.WriteFile(tmpDir+"/main.go", []byte(code), 0644); err != nil {
		return "", fmt.Errorf("error creating temp file %q: %v", tmpDir, err)
	}

	exe := filepath.Join(tmpDir, "a.out")
	goCache := filepath.Join(tmpDir, "gocache")
	cmd := exec.Command("go", "build", "-C", tmpDir, "-o", exe)
	var goPath string
	// cmd.Env = []string{"GOOS=nacl", "GOARCH=amd64p32", "GOCACHE=" + goCache}
	cmd.Env = []string{"GOCACHE=" + goCache}

	useModules := false
	if useModules {
		// Create a GOPATH just for modules to be downloaded
		// into GOPATH/pkg/mod.
		goPath, err = os.MkdirTemp("", "gopath")
		if err != nil {
			return "", fmt.Errorf("error creating temp directory: %v", err)
		}
		defer os.RemoveAll(goPath)
		cmd.Env = append(cmd.Env, "GO111MODULE=on", "GOPROXY=https://proxy.golang.org")
	} else {
		goPath = os.Getenv("GOPATH")
		cmd.Env = append(cmd.Env, "GO111MODULE=off")
	}
	cmd.Env = append(cmd.Env, "GOPATH="+goPath)

	if out, err := cmd.CombinedOutput(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			errs := strings.Replace(string(out), tmpDir+"/", "", -1)
			errs = strings.Replace(errs, "# command-line-arguments\n", "", 1)
			return "", errors.New("errors: " + errs)
		}
		return "", fmt.Errorf("error building go source: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	cmd = exec.CommandContext(ctx, exe)
	var recOut bytes.Buffer
	var recErr bytes.Buffer
	cmd.Stdout = &recOut
	cmd.Stderr = &recErr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("process took too long. out: %s, err: %s", recOut.String(), recErr.String())
		}
	}
	return recOut.String(), nil
}
