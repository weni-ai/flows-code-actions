package coderunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var resourceConfig *specs.LinuxResources

type Service struct {
	codeRun *coderun.Service
	codeLog *codelog.Service
	confs   *config.Config
}

func NewCodeRunnerService(confs *config.Config, coderun *coderun.Service, codelog *codelog.Service) *Service {
	return &Service{codeRun: coderun, codeLog: codelog, confs: confs}
}

func (s *Service) RunCode(ctx context.Context, codeID string, code string, language string, params map[string]interface{}, body string) (*coderun.CodeRun, error) {
	cID, _ := primitive.ObjectIDFromHex(codeID)
	cr := &coderun.CodeRun{
		CodeID: cID,
		Status: coderun.StatusStarted,
		Params: params,
		Body:   body,
	}

	newCodeRun, err := s.codeRun.Create(ctx, cr)
	if err != nil {
		return nil, err
	}

	switch language {
	case "python":
		_, err = s.runPython(ctx, codeID, newCodeRun.ID.Hex(), code, params, body)
	case "javascript":
		_, err = runJs(ctx, code)
	case "go":
		_, err = runGo(ctx, code)
	default:
		return nil, errors.New("unsupported language code type")
	}
	if err != nil {
		log.WithError(err).Error(err.Error())
		newCodeRun.Status = coderun.StatusFailed
		newCodeRun.Result = errors.Wrap(err, "error on executing code").Error()
		errcoderun, cerr := s.codeRun.Update(ctx, newCodeRun.ID.Hex(), newCodeRun)
		if cerr != nil {
			return errcoderun, cerr
		}
		return errcoderun, errors.Wrap(err, "error on executing code")
	}

	newCodeRun, err = s.codeRun.GetByID(ctx, newCodeRun.ID.Hex())
	if err != nil {
		return nil, err
	}
	newCodeRun.Status = coderun.StatusCompleted
	return s.codeRun.Update(ctx, newCodeRun.ID.Hex(), newCodeRun)
}

var environment = ""

func init() {
	environment = config.Getenv("FLOWS_CODE_ACTIONS_ENVIRONMENT", "local")
}

func (s *Service) runPython(ctx context.Context, codeID string, coderunID string, code string, params map[string]interface{}, body string) (string, error) {
	tempDir, err := os.MkdirTemp("./", "code-")
	if err != nil {
		fmt.Println("Error ao criar diretório temporário:", err)
		return "", err
	}
	defer os.RemoveAll(tempDir)

	//TODO: figure out how to handle temporary files dir
	currentDir := "/home/rafaelsoares/weni/weni-ai/codeactions"
	if environment != "local" {
		currentDir = "/app"
	}
	sourceFile := currentDir + "/engines/py/main.py"
	destinatinFile := tempDir + "/main.py"
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		return "", errors.Wrap(err, "Error on reading main file")
	}
	err = os.WriteFile(destinatinFile, data, 0644)
	if err != nil {
		return "", errors.Wrap(err, "Error on copy main file")
	}

	codeFile := tempDir + "/action.py"
	err = os.WriteFile(codeFile, []byte(code), 0644)
	if err != nil {
		return "", errors.Wrap(err, "Error on create code file")
	}

	paramsArgs := ""
	if len(params) > 0 {
		paramsArgs = "-a "
		paramsjs, err := json.Marshal(params)
		if err != nil {
			return "", err
		}
		paramsArgs += string(paramsjs)
	}

	bodyArg := ""
	if body != "" {
		bodyArg = fmt.Sprintf("-b %s", body)
	}

	idRunArg := fmt.Sprintf("-r %s", coderunID)
	idCodeArg := fmt.Sprintf("-c %s", codeID)

	var cmd *exec.Cmd
	if paramsArgs != "" {
		if bodyArg != "" {
			cmd = exec.Command("python", tempDir+"/main.py", paramsArgs, bodyArg, idRunArg, idCodeArg)
		} else {
			cmd = exec.Command("python", tempDir+"/main.py", paramsArgs, idRunArg, idCodeArg)
		}
	} else {
		if bodyArg != "" {
			cmd = exec.Command("python", tempDir+"/main.py", bodyArg, idRunArg, idCodeArg)
		} else {
			cmd = exec.Command("python", tempDir+"/main.py", idRunArg, idCodeArg)
		}
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("process took too long. out: %s, err: %s", stdout.String(), stderr.String())
		}
	}

	if s.confs.ResourceManagement.Enabled {
		cg, err := InitCGroup(ctx, s.confs, codeID)
		if err != nil {
			return "", err
		}
		cg.AddProc(uint64(cmd.Process.Pid))
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("process took too long. out: %s, err: %s", stdout.String(), stderr.String())
		}
	}
	if stdout.String() != "" {
		log.Println("code run stdout: ", stdout.String())
	}
	if stderr.String() != "" {
		return stderr.String(), fmt.Errorf("error executing code: %s", stderr.String())
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

	if err := os.WriteFile(tmpDir+"/main.go", []byte(code), 0644); err != nil {
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

// InitCGroup load or create a new CGroup for the given code
func InitCGroup(ctx context.Context, config *config.Config, codeID string) (cgroup1.Cgroup, error) {
	cgpath := cgroup1.StaticPath("/" + codeID)
	cg, err := cgroup1.Load(cgpath)
	if err != nil {
		resources := ResourceConfig(config)
		cg, err = cgroup1.New(
			cgpath,
			resources,
		)
		if err != nil {
			err = errors.Wrap(err, "failed to create cgroup")
			log.Error(err)
			return nil, err
		}
	}
	return cg, nil
}

// ResourceConfig returns a resource config for the given application configuration
func ResourceConfig(config *config.Config) *specs.LinuxResources {
	if resourceConfig != nil {
		return resourceConfig
	}
	cpu := &specs.LinuxCPU{
		Shares: config.ResourceManagement.CPU.Shares,
		Quota:  config.ResourceManagement.CPU.Quota,
	}
	memory := &specs.LinuxMemory{
		Limit:       config.ResourceManagement.Memory.Limit,
		Reservation: config.ResourceManagement.Memory.Reservation,
	}
	resourceConfig = &specs.LinuxResources{
		CPU:    cpu,
		Memory: memory,
	}
	return resourceConfig
}
