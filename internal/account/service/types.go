package service

import (
	"context"

	"github.com/tsukiyo/mercury/internal/account/domain"
)

type AccountService interface {
	Credit(ctx context.Context, credit domain.Credit) error
}
