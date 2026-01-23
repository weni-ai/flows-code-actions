package workerpool

import (
	"context"
	"errors"

	"github.com/weni-ai/flows-code-actions/internal/coderun"
)

var errNoExecutor = errors.New("workerpool task has no executor")

type Result struct {
	Run *coderun.CodeRun
	Err error
}

type Task struct {
	Ctx     context.Context
	Execute func(ctx context.Context) (*coderun.CodeRun, error)
	Result  chan Result
}

type Pool struct {
	tasks chan Task
}

func NewPool(workers int, queueSize int) *Pool {
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = workers
	}

	pool := &Pool{
		tasks: make(chan Task, queueSize),
	}

	for i := 0; i < workers; i++ {
		go pool.worker()
	}

	return pool
}

func (p *Pool) Submit(task Task) error {
	if task.Ctx != nil {
		select {
		case <-task.Ctx.Done():
			return task.Ctx.Err()
		default:
		}
	}

	select {
	case p.tasks <- task:
		return nil
	default:
		return errors.New("workerpool queue is full")
	}
}

func (p *Pool) worker() {
	for task := range p.tasks {
		if task.Execute == nil {
			if task.Result != nil {
				task.Result <- Result{Err: errNoExecutor}
			}
			continue
		}

		if task.Ctx != nil {
			select {
			case <-task.Ctx.Done():
				if task.Result != nil {
					task.Result <- Result{Err: task.Ctx.Err()}
				}
				continue
			default:
			}
		}

		run, err := task.Execute(task.Ctx)
		if task.Result != nil {
			task.Result <- Result{Run: run, Err: err}
		}
	}
}
