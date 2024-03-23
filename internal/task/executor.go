package task

import (
	"context"
	"fmt"

	"github.com/tsukaychan/webook/internal/domain"
)

type Executor interface {
	Name() string
	Exec(ctx context.Context, task domain.Task) error
}

type LocalFuncExecutor struct {
	fns map[string]func(ctx context.Context, tsk domain.Task) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		fns: make(map[string]func(ctx context.Context, tsk domain.Task) error),
	}
}

func (executor *LocalFuncExecutor) AddLocalFunc(name string, fn func(ctx context.Context, tsk domain.Task) error) {
	executor.fns[name] = fn
}

func (executor *LocalFuncExecutor) Name() string {
	return "local"
}

func (executor *LocalFuncExecutor) Exec(ctx context.Context, task domain.Task) error {
	fn, ok := executor.fns[task.Name]
	if !ok {
		return fmt.Errorf("unknown task or unregistered: %v", task.Name)
	}

	return fn(ctx, task)
}
