package service

import (
	"context"

	"github.com/tsukiyo/mercury/internal/account/domain"
	"github.com/tsukiyo/mercury/internal/account/repository"
)

type accountService struct {
	repo repository.AccountRepository
}

func NewAccountServiceServer(repo repository.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

func (a *accountService) Credit(ctx context.Context, credit domain.Credit) error {
	return a.repo.AddCredit(ctx, credit)
}
