package workerpool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/weni-ai/flows-code-actions/internal/coderun"
)

func waitResult(t *testing.T, ch <-chan Result) Result {
	t.Helper()

	select {
	case res := <-ch:
		return res
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for workerpool result")
		return Result{}
	}
}

func waitSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for workerpool signal")
	}
}

func TestPoolExecutesTask(t *testing.T) {
	pool := NewPool(1, 1)

	resultCh := make(chan Result, 1)
	expectedID := "run-123"
	task := Task{
		Ctx: context.Background(),
		Execute: func(ctx context.Context) (*coderun.CodeRun, error) {
			return &coderun.CodeRun{ID: expectedID}, nil
		},
		Result: resultCh,
	}

	if err := pool.Submit(task); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	res := waitResult(t, resultCh)
	if res.Err != nil {
		t.Fatalf("unexpected execution error: %v", res.Err)
	}
	if res.Run == nil || res.Run.ID != expectedID {
		t.Fatalf("unexpected result: %#v", res.Run)
	}
}

func TestPoolSubmitContextCanceled(t *testing.T) {
	pool := NewPool(1, 1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	task := Task{
		Ctx: ctx,
		Execute: func(ctx context.Context) (*coderun.CodeRun, error) {
			return &coderun.CodeRun{}, nil
		},
		Result: make(chan Result, 1),
	}

	if err := pool.Submit(task); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestPoolSubmitQueueFull(t *testing.T) {
	pool := NewPool(1, 1)

	block := make(chan struct{})
	started := make(chan struct{})
	task1 := Task{
		Ctx: context.Background(),
		Execute: func(ctx context.Context) (*coderun.CodeRun, error) {
			close(started)
			<-block
			return &coderun.CodeRun{}, nil
		},
		Result: make(chan Result, 1),
	}
	if err := pool.Submit(task1); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}
	waitSignal(t, started)

	task2 := Task{
		Ctx:     context.Background(),
		Execute: func(ctx context.Context) (*coderun.CodeRun, error) { return &coderun.CodeRun{}, nil },
		Result:  make(chan Result, 1),
	}
	if err := pool.Submit(task2); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	task3 := Task{
		Ctx:     context.Background(),
		Execute: func(ctx context.Context) (*coderun.CodeRun, error) { return &coderun.CodeRun{}, nil },
		Result:  make(chan Result, 1),
	}
	if err := pool.Submit(task3); err == nil || err.Error() != "workerpool queue is full" {
		t.Fatalf("expected queue full error, got: %v", err)
	}

	close(block)
}

func TestPoolWorkerNoExecutor(t *testing.T) {
	pool := NewPool(1, 1)

	resultCh := make(chan Result, 1)
	task := Task{
		Ctx:    context.Background(),
		Result: resultCh,
	}

	if err := pool.Submit(task); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	res := waitResult(t, resultCh)
	if res.Err == nil || res.Err.Error() != errNoExecutor.Error() {
		t.Fatalf("expected errNoExecutor, got: %v", res.Err)
	}
}

func TestPoolWorkerContextCanceledBeforeExecute(t *testing.T) {
	pool := NewPool(1, 1)

	block := make(chan struct{})
	started := make(chan struct{})
	task1 := Task{
		Ctx: context.Background(),
		Execute: func(ctx context.Context) (*coderun.CodeRun, error) {
			close(started)
			<-block
			return &coderun.CodeRun{}, nil
		},
		Result: make(chan Result, 1),
	}
	if err := pool.Submit(task1); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}
	waitSignal(t, started)

	ctx, cancel := context.WithCancel(context.Background())
	resultCh := make(chan Result, 1)
	task2 := Task{
		Ctx: ctx,
		Execute: func(ctx context.Context) (*coderun.CodeRun, error) {
			return &coderun.CodeRun{}, nil
		},
		Result: resultCh,
	}
	if err := pool.Submit(task2); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	cancel()
	close(block)

	res := waitResult(t, resultCh)
	if !errors.Is(res.Err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", res.Err)
	}
}
