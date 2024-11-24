package repository

import (
	"context"

	"github.com/lazywoo/mercury/internal/account/domain"
)

type AccountRepository interface {
	AddCredit(ctx context.Context, credit domain.Credit) error
}

func NewAccountRepository() AccountRepository {
	return &accountRepository{}
}
