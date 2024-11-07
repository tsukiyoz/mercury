package validator

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/lazywoo/mercury/pkg/migrator"
	"github.com/lazywoo/mercury/pkg/migrator/events"
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

	// loopInterval is the interval between each synchronization,
	// <= 0 means no sleep, quit this loop,
	// > 0 means sleep this interval.
	// unit is second
	loopInterval time.Duration

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
	l logger.Logger,
) *Validator[T] {
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

// WithLoopInterval set the sleep interval between each synchronization
func (v *Validator[T]) WithLoopInterval(loopInterval time.Duration) *Validator[T] {
	v.loopInterval = max(loopInterval, time.Millisecond*200)
	return v
}

// WithMaxRetry set the max retry times
func (v *Validator[T]) WithMaxRetry(maxRetry int) *Validator[T] {
	v.maxRetry = maxRetry
	return v
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	v.l.Debug("validation started")
	var eg errgroup.Group
	eg.Go(func() error {
		v.baseToTarget(ctx)
		return nil
	})
	eg.Go(func() error {
		v.targetToBase(ctx)
		return nil
	})

	err := eg.Wait()
	v.l.Debug("validate finished")
	return err
}

// validateBaseToTarget try to synchronize data from base to target
func (v *Validator[T]) baseToTarget(ctx context.Context) {
	offset := 0

	for {
		var src T
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		srcErr := v.base.WithContext(dbCtx).
			Where("utime > ?", v.utime).
			Offset(offset).
			Order("id").
			First(&src).Error
		cancel()

		switch srcErr {
		case context.Canceled, context.DeadlineExceeded:
			v.l.Debug("exit base==>target validation")
			return
		case nil:
			// find data
			dbCtx, cancel := context.WithTimeout(ctx, time.Second)
			var dst T
			dstErr := v.target.WithContext(dbCtx).Where("id = ?", src.ID()).First(&dst).Error
			cancel()

			switch dstErr {
			case nil:
				if !src.Equal(dst) {
					v.notify(ctx, src.ID(), events.InconsistentEventTypeNotEqual)
				}
			case gorm.ErrRecordNotFound:
				v.notify(ctx, src.ID(), events.InconsistentEventTypeTargetMissing)
			default:
				v.l.Error("validate data, query target failed", logger.Error(dstErr))
				continue
			}

		case gorm.ErrRecordNotFound:
			// no more data
			if v.loopInterval <= 0 {
				// finish validate
				v.l.Debug("exit base==>target validation")
				return
			}
			time.Sleep(v.loopInterval)
			continue

		default:
			v.l.Error("validate data, query base failed", logger.Error(srcErr))
			continue
		}

		offset++
	}
}

// targetToBase find data from target that doesn't exist in base and try to fix(delete) it
func (v *Validator[T]) targetToBase(ctx context.Context) {
	offset := 0

	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		var dstTs []T
		err := v.target.WithContext(dbCtx).
			Where("utime > ?", v.utime).
			Select("id").
			Offset(offset).
			Limit(v.batchSize).
			First(&dstTs).Error
		cancel()

		switch err {
		case context.Canceled, context.DeadlineExceeded:
			v.l.Debug("exit target==>base validation")
			return

		case nil:
			ids := slice.Map(dstTs, func(idx int, t T) int64 {
				return t.ID()
			})
			var srcTs []T
			dbCtx, cancel := context.WithTimeout(ctx, time.Second)
			err := v.base.WithContext(dbCtx).Where("id IN ?", ids).Find(&srcTs).Error
			cancel()
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
				v.l.Error("validate data, query base failed", logger.Error(err))
				continue
			}

		case gorm.ErrRecordNotFound:
			if v.loopInterval <= 0 {
				v.l.Debug("exit target==>base validation")
				return
			}
			time.Sleep(v.loopInterval)
			continue

		default:
			v.l.Error("validate data, query target failed", logger.Error(err))
			continue
		}

		if len(dstTs) < v.batchSize {
			// no more data
			if v.loopInterval <= 0 {
				return
			}
			time.Sleep(v.loopInterval)
			offset += len(dstTs)
			continue
		}

		offset += v.batchSize
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
