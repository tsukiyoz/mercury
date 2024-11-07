package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	taskStatusUnknown = iota
	taskStatusWaiting
	taskStatusRunning
	taskStatusEnd
)

type Task struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"unique"`
	Executor   string
	Cfg        string
	Expression string
	Version    int64
	NextTime   int64 `gorm:"index"`
	Status     int
	Ctime      int64
	Utime      int64
}

type TaskDAO interface {
	Preempt(ctx context.Context) (Task, error)
	Release(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, nxt time.Time) error
	UpdateUtime(ctx context.Context, id int64) error
	Insert(ctx context.Context, tsk Task) error
}

type GORMTaskDAO struct {
	db *gorm.DB
}

func NewGORMTaskDAO(db *gorm.DB) TaskDAO {
	return &GORMTaskDAO{
		db: db,
	}
}

func (dao *GORMTaskDAO) Preempt(ctx context.Context) (Task, error) {
	db := dao.db.WithContext(ctx)
	for {
		now := time.Now().UnixMilli()
		// get task
		var task Task
		err := db.WithContext(ctx).
			Model(&Task{}).
			Where("next_time <= ? AND status = ?", now, taskStatusWaiting).
			First(&task).Error
		if err != nil {
			return Task{}, err
		}

		// preempt
		res := db.WithContext(ctx).
			Model(&Task{}).
			Where("id = ? AND version = ?", task.Id, task.Version).
			Updates(map[string]any{
				"utime":   now,
				"version": task.Version + 1,
				"status":  taskStatusRunning,
			})
		if res.Error != nil {
			return Task{}, err
		}
		// preempt success
		if res.RowsAffected == 1 {
			return task, nil
		}
	}
}

func (dao *GORMTaskDAO) Release(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	// TODO query with version field
	return dao.db.WithContext(ctx).
		Model(&Task{}).
		Where("id = ? AND status = ?", id, taskStatusRunning).
		Updates(map[string]any{
			"utime":  now,
			"status": taskStatusWaiting,
		}).Error
}

func (dao *GORMTaskDAO) UpdateNextTime(ctx context.Context, id int64, nxt time.Time) error {
	return dao.db.WithContext(ctx).
		Model(&Task{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"utime":     time.Now().UnixMilli(),
			"next_time": nxt.UnixMilli(),
		}).Error
}

func (dao *GORMTaskDAO) UpdateUtime(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).
		Model(&Task{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"utime": time.Now().UnixMilli(),
		}).Error
}

func (dao *GORMTaskDAO) Insert(ctx context.Context, tsk Task) error {
	now := time.Now().UnixMilli()
	tsk.Ctime, tsk.Utime = now, now
	return dao.db.WithContext(ctx).Create(&tsk).Error
}
