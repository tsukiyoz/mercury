package repository

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/lazywoo/mercury/internal/account/domain"
	"github.com/lazywoo/mercury/internal/account/repository/dao"
)

type accountRepository struct {
	dao dao.AccountDAO
}

func (a *accountRepository) AddCredit(ctx context.Context, credit domain.Credit) error {
	now := time.Now().UnixMilli()
	activities := lo.Map(credit.Items, func(v domain.CreditItem, _ int) dao.AccountActivity {
		return dao.AccountActivity{
			Uid:         v.Uid,
			Biz:         credit.Biz,
			BizId:       credit.BizId,
			Account:     v.Account,
			AccountType: v.AccountType.AsUint8(),
			Amount:      int64(v.Amount),
			Currency:    v.Currency,
			Ctime:       now,
			Utime:       now,
		}
	})
	return a.dao.AddActivities(ctx, activities)
}
