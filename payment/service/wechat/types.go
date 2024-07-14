package wechat

import (
	"context"
	"github.com/lazywoo/mercury/payment/domain"
)

type PaymentService interface {
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}
