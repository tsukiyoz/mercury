package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountGORMDAO struct {
	db *gorm.DB
}

func NewAccountDAO(db *gorm.DB) AccountDAO {
	return &AccountGORMDAO{db: db}
}

func (g *AccountGORMDAO) AddActivities(ctx context.Context, activities []AccountActivity) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		for _, activiy := range activities {
			err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.Assignments(map[string]any{
					"balance": gorm.Expr("balance + ?", activiy.Amount),
					"utime":   now,
				}),
			}).Create(&Account{
				Uid:      activiy.Uid,
				Account:  activiy.Account,
				Type:     activiy.AccountType,
				Balance:  int(activiy.Amount),
				Currency: activiy.Currency,
				Ctime:    now,
				Utime:    now,
			}).Error
			if err != nil {
				return err
			}
		}
		return tx.Create(&activities).Error
	})
}
