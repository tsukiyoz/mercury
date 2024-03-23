package repository

import (
	"context"
	"time"

	"github.com/tsukaychan/webook/internal/repository/dao"

	"github.com/tsukaychan/webook/internal/domain"
)

type TaskRepository interface {
	Preempt(ctx context.Context) (domain.Task, error)
	Release(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
	UpdateUtime(ctx context.Context, id int64) error
	AddJob(ctx context.Context, tsk domain.Task) error
}

var _ TaskRepository = (*PreemptTaskRepository)(nil)

type PreemptTaskRepository struct {
	dao dao.TaskDAO
}

func (repo *PreemptTaskRepository) Preempt(ctx context.Context) (domain.Task, error) {
	task, err := repo.dao.Preempt(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	return repo.entityToDomain(task), nil
}

func (repo *PreemptTaskRepository) Release(ctx context.Context, id int64) error {
	return repo.dao.Release(ctx, id)
}

func (repo *PreemptTaskRepository) UpdateNextTime(ctx context.Context, id int64, t time.Time) error {
	return repo.dao.UpdateNextTime(ctx, id, t)
}

func (repo *PreemptTaskRepository) UpdateUtime(ctx context.Context, id int64) error {
	return repo.dao.UpdateUtime(ctx, id)
}

func (repo *PreemptTaskRepository) AddJob(ctx context.Context, tsk domain.Task) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(tsk))
}

func (repo *PreemptTaskRepository) entityToDomain(tsk dao.Task) domain.Task {
	return domain.Task{
		Id:         tsk.Id,
		Name:       tsk.Name,
		Executor:   tsk.Executor,
		Cfg:        tsk.Executor,
		Expression: tsk.Expression,
		NextTime:   time.UnixMilli(tsk.NextTime),
	}
}

func (repo *PreemptTaskRepository) domainToEntity(tsk domain.Task) dao.Task {
	return dao.Task{
		Id:         tsk.Id,
		Name:       tsk.Name,
		Executor:   tsk.Executor,
		Cfg:        tsk.Executor,
		Expression: tsk.Expression,
		NextTime:   tsk.NextTime.UnixMilli(),
	}
}
