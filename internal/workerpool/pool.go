package workerpool

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/weni-ai/flows-code-actions/internal/coderun"
	"github.com/weni-ai/flows-code-actions/internal/metrics"
)

var errNoExecutor = errors.New("workerpool task has no executor")

type Result struct {
	Run *coderun.CodeRun
	Err error
}

type Task struct {
	Ctx      context.Context
	Execute  func(ctx context.Context) (*coderun.CodeRun, error)
	Result   chan Result
	queuedAt time.Time // timestamp when task entered the queue
}

type Pool struct {
	tasks       chan Task
	workerCount int
	queueSize   int
	busyWorkers int64 // atomic counter for busy workers
}

func NewPool(workers int, queueSize int) *Pool {
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = workers
	}

	pool := &Pool{
		tasks:       make(chan Task, queueSize),
		workerCount: workers,
		queueSize:   queueSize,
	}

	// Register capacity metrics
	metrics.SetWorkerpoolWorkersTotal(float64(workers))
	metrics.SetWorkerpoolQueueCapacity(float64(queueSize))

	for i := 0; i < workers; i++ {
		go pool.worker()
	}

	return pool
}

func (p *Pool) Submit(task Task) error {
	if task.Ctx != nil {
		select {
		case <-task.Ctx.Done():
			metrics.IncWorkerpoolTasksTimeout()
			return task.Ctx.Err()
		default:
		}
	}

	task.queuedAt = time.Now() // Mark queue entry time

	select {
	case p.tasks <- task:
		metrics.IncWorkerpoolTasksSubmitted()
		metrics.SetWorkerpoolQueueSize(float64(len(p.tasks)))
		return nil
	default:
		metrics.IncWorkerpoolTasksRejected()
		return errors.New("workerpool queue is full")
	}
}

func (p *Pool) worker() {
	for task := range p.tasks {
		// Update queue size metric
		metrics.SetWorkerpoolQueueSize(float64(len(p.tasks)))

		// Measure queue wait time
		if !task.queuedAt.IsZero() {
			queueWait := time.Since(task.queuedAt).Seconds()
			metrics.ObserveWorkerpoolQueueWait(queueWait)
		}

		if task.Execute == nil {
			if task.Result != nil {
				task.Result <- Result{Err: errNoExecutor}
			}
			metrics.IncWorkerpoolTasksFailed()
			continue
		}

		if task.Ctx != nil {
			select {
			case <-task.Ctx.Done():
				if task.Result != nil {
					task.Result <- Result{Err: task.Ctx.Err()}
				}
				metrics.IncWorkerpoolTasksTimeout()
				continue
			default:
			}
		}

		// Increment busy workers
		atomic.AddInt64(&p.busyWorkers, 1)
		metrics.IncWorkerpoolWorkersBusy()

		// Execute and measure duration
		execStart := time.Now()
		run, err := task.Execute(task.Ctx)
		execDuration := time.Since(execStart).Seconds()

		// Decrement busy workers
		atomic.AddInt64(&p.busyWorkers, -1)
		metrics.DecWorkerpoolWorkersBusy()

		// Record result metrics
		metrics.ObserveWorkerpoolTaskDuration(execDuration)
		if err != nil {
			metrics.IncWorkerpoolTasksFailed()
		} else {
			metrics.IncWorkerpoolTasksCompleted()
		}

		if task.Result != nil {
			task.Result <- Result{Run: run, Err: err}
		}
	}
}

// GetStats returns current pool statistics (useful for debugging)
func (p *Pool) GetStats() (workers, busy, queueLen, queueCap int) {
	return p.workerCount, int(atomic.LoadInt64(&p.busyWorkers)), len(p.tasks), p.queueSize
}
