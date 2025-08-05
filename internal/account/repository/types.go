package repository

import (
	"context"

	"github.com/tsukiyo/mercury/internal/account/domain"
	"github.com/tsukiyo/mercury/internal/account/repository/dao"
)

type AccountRepository interface {
	AddCredit(ctx context.Context, credit domain.Credit) error
}

func NewAccountRepository(dao dao.AccountDAO) AccountRepository {
	return &accountRepository{dao: dao}
}
