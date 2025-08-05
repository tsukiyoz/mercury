package grpc

import (
	"context"

	"github.com/samber/lo"
	"google.golang.org/grpc"

	accountv1 "github.com/tsukiyo/mercury/api/gen/account/v1"
	"github.com/tsukiyo/mercury/internal/account/domain"
	"github.com/tsukiyo/mercury/internal/account/service"
)

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer
	svc service.AccountService
}

func NewAccountServiceServer(svc service.AccountService) *AccountServiceServer {
	return &AccountServiceServer{svc: svc}
}

func (a *AccountServiceServer) Register(server grpc.ServiceRegistrar) {
	accountv1.RegisterAccountServiceServer(server, a)
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
