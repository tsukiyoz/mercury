package validator

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/tsukaychan/mercury/pkg/logger"
	"github.com/tsukaychan/mercury/pkg/migrator"
	"github.com/tsukaychan/mercury/pkg/migrator/events"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

type Validator[T migrator.Entity] struct {
	base     *gorm.DB
	target   *gorm.DB
	producer events.InconsistentEventProducer

	// direction is the direction of the validation
	direction migrator.Direction

	// batchSize is the batch size of the data
	// in targetToBase func.
	// default is 100
	batchSize int

	// utime is the utime of the last synchronized data
	utime int64

	// sleepInterval is the interval between each synchronization,
	// <= 0 means no sleep, quit this loop,
	// > 0 means sleep this interval.
	// unit is second
	sleepInterval time.Duration

	// order is the order of the data
	order string

	// maxRetry is the max retry times
	// default is 3
	maxRetry int

	l logger.Logger
}

func NewValidator[T migrator.Entity](
	base, target *gorm.DB,
	producer events.InconsistentEventProducer,
	direction migrator.Direction,
	l logger.Logger) *Validator[T] {
	return &Validator[T]{
		base:      base,
		target:    target,
		producer:  producer,
		direction: direction,
		l:         l,
		batchSize: 100,
		maxRetry:  3,
	}
}

// WithBatchSize set the batch size of the data
func (v *Validator[T]) WithBatchSize(batchSize int) *Validator[T] {
	v.batchSize = batchSize
	return v
}

// WithUtime set the utime of the validation
func (v *Validator[T]) WithUtime(utime int64) *Validator[T] {
	v.utime = utime
	return v
}

// WithSleepInterval set the sleep interval between each synchronization
func (v *Validator[T]) WithSleepInterval(sleepInterval time.Duration) *Validator[T] {
	v.sleepInterval = sleepInterval
	return v
}

// WithMaxRetry set the max retry times
func (v *Validator[T]) WithMaxRetry(maxRetry int) *Validator[T] {
	v.maxRetry = maxRetry
	return v
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		v.baseToTarget(ctx)
		return nil
	})
	eg.Go(func() error {
		v.targetToBase(ctx)
		return nil
	})

	return eg.Wait()
}

// validateBaseToTarget try to synchronize data from base to target
func (v *Validator[T]) baseToTarget(ctx context.Context) {
	offset := 0
	errCnt := 0
	skip := func() bool {
		errCnt++
		if errCnt >= v.maxRetry {
			// skip this data, update offset
			offset++
			errCnt = 0
			return true
		}
		return false
	}

	for {
		var src T
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		srcErr := v.base.WithContext(ctx).
			Where("utime > ?", v.utime).
			Offset(offset).
			Order("utime").
			First(&src).Error
		cancel()

		switch srcErr {
		case nil:
			// find data
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			var dst T
			dstErr := v.target.WithContext(ctx).Where("id = ?", src.ID()).First(&dst).Error
			cancel()

			switch dstErr {
			case nil:
				if !src.Equal(dst) {
					v.notify(ctx, src.ID(), events.InconsistentEventTypeNotEqual)
				}
			case gorm.ErrRecordNotFound:
				v.notify(ctx, src.ID(), events.InconsistentEventTypeTargetMissing)
			default:
				// database error
				if !skip() {
					v.l.Error("validate data, query target failed", logger.Error(dstErr))
					continue
				} else {
					//  skip this data
					v.l.Error(fmt.Sprintf("validate data, query target failed in %s tries", v.maxRetry), logger.Error(dstErr))
				}
			}

		case gorm.ErrRecordNotFound:
			// no more data
			if v.sleepInterval <= 0 {
				// finish validate
				return
			}
			time.Sleep(v.sleepInterval)
			continue

		default:
			// database error
			// retry
			if !skip() {
				v.l.Error("validate data, query base failed", logger.Error(srcErr))
				continue
			} else {
				//  skip this data
				v.l.Error(fmt.Sprintf("validate data, query base failed in %s tries", v.maxRetry), logger.Error(srcErr))
			}
		}

		offset++
		errCnt = 0
	}
}

// targetToBase find data from target that doesn't exist in base and try to fix(delete) it
func (v *Validator[T]) targetToBase(ctx context.Context) {
	offset := 0
	errCnt := 0
	skip := func() bool {
		errCnt++
		if errCnt >= v.maxRetry {
			// skip this data, update offset
			offset++
			errCnt = 0
			return true
		}
		return false
	}

	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		var dstTs []T
		err := v.target.WithContext(dbCtx).
			Where("utime > ?", v.utime).
			Offset(offset).
			Limit(v.batchSize).
			Order("utime").
			First(&dstTs).Error
		cancel()

		if len(dstTs) == 0 {
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue
		}

		switch err {
		case nil:
			ids := slice.Map(dstTs, func(idx int, t T) int64 {
				return t.ID()
			})
			var srcTs []T
			err := v.base.Where("id IN ?", ids).Find(&srcTs).Error
			switch err {
			case nil:
				srcIds := slice.Map(srcTs, func(idx int, t T) int64 {
					return t.ID()
				})
				diff := slice.DiffSet(ids, srcIds)
				v.notifyBaseMissing(ctx, diff)
			case gorm.ErrRecordNotFound:
				v.notifyBaseMissing(ctx, ids)
			default:
				if !skip() {
					v.l.Error("validate data, query base failed", logger.Error(err))
					continue
				}
				v.l.Error(fmt.Sprintf("validate data, query base failed in %s tries", v.maxRetry), logger.Error(err))
			}

		case gorm.ErrRecordNotFound:
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue

		default:
			if !skip() {
				v.l.Error("validate data, query target failed", logger.Error(err))
				continue
			}
			v.l.Error("validate data, query target failed")
		}

		if len(dstTs) < v.batchSize {
			// no more data
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue
		}

		offset += v.batchSize
		errCnt = 0
	}
}

// notify send events to message queue
func (v *Validator[T]) notify(ctx context.Context, id int64, typ events.InconsistentEventType) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Direction: v.direction,
		Type:      typ,
	})
	if err != nil {
		v.l.Error("produce inconsistent event failed", logger.Error(err))
	}
	cancel()
}

func (v *Validator[T]) notifyBaseMissing(ctx context.Context, ids []int64) {
	for _, id := range ids {
		v.notify(ctx, id, events.InconsistentEventTypeBaseMissing)
	}
}
