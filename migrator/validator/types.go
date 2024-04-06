package validator

import (
	"context"
	"github.com/tsukaychan/mercury/migrator"
	"github.com/tsukaychan/mercury/pkg/logger"
	"gorm.io/gorm"
)

type Validator[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	direction string

	batchSize int

	l logger.Logger
}

func (v *Validator[T]) Validate(ctx context.Context) {
	offset := -1
	for {
		offset++
		var t T
		err := v.base.Offset(offset).Order("id").First(&t).Error
		switch err {
		case nil:
			v.target.Where("id = ?", t.ID())
		case gorm.ErrRecordNotFound:
			return
		default:
			v.l.Error("validate date, query base failed", logger.Error(err))
			continue
		}
	}
}
