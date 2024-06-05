package fixer

import (
	"context"
	"github.com/lazywoo/mercury/pkg/migrator"
	"github.com/lazywoo/mercury/pkg/migrator/events"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OverrideFixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

func NewOverrideFixer[T migrator.Entity](base, target *gorm.DB) (*OverrideFixer[T], error) {
	var t T
	rows, err := target.Model(&t).Limit(0).Rows()
	if err != nil {
		return nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: columns,
	}, nil
}

//func (o *OverrideFixer[T]) Fix(ctx context.Context, evt events.InconsistentEvent) error {
//	switch evt.Type {
//	case events.InconsistentEventTypeBaseMissing:
//		// DEL
//		return o.base.WithContext(ctx).Delete(new(T)).Error
//	case events.InconsistentEventTypeTargetMissing,
//		events.InconsistentEventTypeNotEqual:
//		// INSERT
//		var t T
//		err := o.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
//		switch err {
//		case gorm.ErrRecordNotFound:
//			// data has been deleted in base
//			return o.target.WithContext(ctx).Where("id = ?", evt.ID).Delete(&t).Error
//		case nil:
//			// do insert
//			return o.target.Clauses(clause.OnConflict{
//				DoUpdates: clause.AssignmentColumns(o.columns),
//			}).Create(&t).Error
//		default:
//			return err
//		}
//	default:
//		return errors.New("unknown inconsistent_type")
//	}
//}

func (o *OverrideFixer[T]) Fix(ctx context.Context, evt events.InconsistentEvent) error {
	var t T
	err := o.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
	switch err {
	case gorm.ErrRecordNotFound:
		return o.target.WithContext(ctx).Where("id = ?", evt.ID).Delete(&t).Error
	case nil:
		return o.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(o.columns),
		}).Create(&t).Error
	default:
		return err
	}
}
