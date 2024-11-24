package grpc

import (
	"context"

	"github.com/lazywoo/mercury/internal/account/domain"
	"github.com/lazywoo/mercury/internal/account/service"
	accountv1 "github.com/lazywoo/mercury/pkg/api/account/v1"
	"github.com/samber/lo"
)

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer
	svc service.AccountService
}

func NewAccountServiceServer(svc service.AccountService) *AccountServiceServer {
	return &AccountServiceServer{svc: svc}
}

func (a *AccountServiceServer) Credit(ctx context.Context, credit *accountv1.CreditRequest) (*accountv1.CreditResponse, error) {
	err := a.svc.Credit(ctx, domain.Credit{
		Biz:   credit.Biz,
		BizId: credit.BizId,
		Items: lo.Map(credit.Items, func(v *accountv1.CreditItem, idx int) domain.CreditItem {
			return domain.CreditItem{
				Account:     v.Account,
				AccountType: domain.AccountType(v.AccountType),
				Amount:      int(v.Amount),
				Currency:    v.Currency,
				Uid:         v.Uid,
			}
		}),
	})
	return &accountv1.CreditResponse{}, err
}
