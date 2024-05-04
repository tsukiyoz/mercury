package task

import (
	"context"
	"fmt"
	"time"

	"github.com/tsukaychan/mercury/crontask/domain"
	"github.com/tsukaychan/mercury/crontask/service"

	"golang.org/x/sync/semaphore"

	"github.com/tsukaychan/mercury/pkg/logger"
)

// this kind of task scheduling framework is based on mysql
// scheduler get tasks from task service
// and forward to task's executor for executing

// Scheduler preempting tasks and scheduling
type Scheduler struct {
	executors map[string]Executor
	interval  time.Duration
	svc       service.TaskService
	limiter   *semaphore.Weighted
	l         logger.Logger
}

func NewScheduler(svc service.TaskService, l logger.Logger) *Scheduler {
	return &Scheduler{
		executors: make(map[string]Executor, 8),
		interval:  time.Second,
		svc:       svc,
		limiter:   semaphore.NewWeighted(200),
		l:         l,
	}
}

func (s *Scheduler) RegisterTask(ctx context.Context, tsk domain.Task) error {
	return s.svc.AddTask(ctx, tsk)
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Start(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		// preempt task
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		tsk, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			s.l.Error("preempt task failed", logger.Error(err))
			time.Sleep(s.interval)
			continue
		}

		// get executor
		executor, ok := s.executors[tsk.Executor]
		if !ok {
			s.l.Error(fmt.Sprintf("unknown executor or unregisterd: %v", tsk.Executor))
			continue
		}

		// execute task in goroutine
		go func() {
			defer func() {
				s.limiter.Release(1)
				if tsk.CancelFunc != nil {
					tsk.CancelFunc()
				}
			}()

			if err := executor.Exec(ctx, tsk); err != nil {
				s.l.Error("task execute failed", logger.Error(err))
				return
			}

			innerCtx, cancel := context.WithTimeout(ctx, time.Second)
			if err := s.svc.ResetNextTime(innerCtx, tsk); err != nil {
				s.l.Error("reset next task time failed", logger.Error(err))
			}
			cancel()
		}()
	}
}
