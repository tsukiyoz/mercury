package service

import (
	"context"

	"github.com/lazywoo/mercury/internal/account/domain"
)

type AccountService interface {
	Credit(ctx context.Context, credit domain.Credit) error
}
