package validator

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/tsukaychan/mercury/migrator"
	"github.com/tsukaychan/mercury/migrator/events"
	"github.com/tsukaychan/mercury/pkg/logger"
	"gorm.io/gorm"
	"time"
)

type Validator[T migrator.Entity] struct {
	base     *gorm.DB
	target   *gorm.DB
	producer events.InconsistentEventProducer

	direction DirectionType

	batchSize int

	l logger.Logger
}

func NewValidator[T migrator.Entity](base,
	target *gorm.DB,
	producer events.InconsistentEventProducer,
	direction DirectionType,
	l logger.Logger) *Validator[T] {
	return &Validator[T]{
		base:      base,
		target:    target,
		producer:  producer,
		direction: direction,
		l:         l,
	}
}

type DirectionType uint

const (
	DirectionTypeToBase DirectionType = iota
	DirectionTypeToTarget
)

func (v *Validator[T]) Validate(ctx context.Context) {
	v.validateBaseToTarget(ctx)
	v.validateTargetToBase(ctx)
}

func (v *Validator[T]) validateBaseToTarget(ctx context.Context) {
	offset := -1
	for {
		offset++

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		var src T
		srcErr := v.base.WithContext(ctx).Offset(offset).Order("id").First(&src).Error
		cancel()

		switch srcErr {
		case nil:
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
				v.l.Error("validate data, query target failed", logger.Error(dstErr))
				continue
			}
		case gorm.ErrRecordNotFound:
			// finished
			return
		default:
			v.l.Error("validate date, query base failed", logger.Error(srcErr))
			continue
		}
	}
}

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

func (v *Validator[T]) validateTargetToBase(ctx context.Context) {
	offset := -v.batchSize
	for {
		offset += v.batchSize
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		var dstTs []T
		err := v.target.WithContext(dbCtx).
			Offset(offset).
			Limit(v.batchSize).
			Order("id").
			First(&dstTs).Error
		cancel()
		if len(dstTs) == 0 {
			return
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
				continue
			}
		case gorm.ErrRecordNotFound:
			return
		default:
			v.l.Error("")
			continue
		}
		if len(dstTs) < v.batchSize {
			return
		}
	}
}
