package service

import (
	"context"
	"time"

	"github.com/lazywoo/mercury/internal/crontask/domain"
	"github.com/lazywoo/mercury/internal/crontask/repository"

	"github.com/lazywoo/mercury/pkg/logger"
)

type TaskService interface {
	Preempt(ctx context.Context) (domain.Task, error)
	ResetNextTime(ctx context.Context, tsk domain.Task) error
	AddTask(ctx context.Context, tsk domain.Task) error
}

var _ TaskService = (*taskService)(nil)

type taskService struct {
	repo                   repository.TaskRepository
	l                      logger.Logger
	refreshInterval        time.Duration
	refreshMaxFailureCount int
}

func NewTaskService(repo repository.TaskRepository, l logger.Logger) TaskService {
	return &taskService{
		repo: repo,
		l:    l,
	}
}

func (svc *taskService) Preempt(ctx context.Context) (domain.Task, error) {
	tsk, err := svc.repo.Preempt(ctx)
	if err != nil {
		return domain.Task{}, err
	}

	ticker := time.NewTicker(svc.refreshInterval)
	go func() {
		failedCnt := 0
		for range ticker.C {
			if svc.refresh(tsk.Id) != nil {
				failedCnt++
			} else {
				failedCnt = 0
			}

			if failedCnt == svc.refreshMaxFailureCount {
				// reckon that the contract renewal has failed
				// TODO notify the scheduler to cancel task execution
				return
			}
		}
	}()

	tsk.CancelFunc = func() {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := svc.repo.Release(ctx, tsk.Id)
		if er != nil {
			svc.l.Error("release tsk failed",
				logger.Error(er),
				logger.Int64("id", tsk.Id),
			)
		}
	}
	return tsk, nil
}

func (svc *taskService) refresh(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := svc.repo.UpdateUtime(ctx, id)
	if err != nil {
		svc.l.Error("renewal failed",
			logger.Int64("id", id),
			logger.Error(err),
		)
		return err
	}
	return nil
}

func (svc *taskService) ResetNextTime(ctx context.Context, tsk domain.Task) error {
	nxt := tsk.GetNextTime(time.Now())
	if nxt.IsZero() {
		return nil
	}
	return svc.repo.UpdateNextTime(ctx, tsk.Id, nxt)
}

func (svc *taskService) AddTask(ctx context.Context, tsk domain.Task) error {
	return svc.repo.AddJob(ctx, tsk)
}
