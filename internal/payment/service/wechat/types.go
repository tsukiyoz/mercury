package wechat

import (
	"context"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"

	"github.com/lazywoo/mercury/internal/payment/domain"
)

type PaymentService interface {
	Prepay(ctx context.Context, pmt domain.Payment) (string, error) // 预支付
	SyncInfo(ctx context.Context, bizTradeNo string) error
	FindExpiredPayments(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error)
	GetPayment(ctx context.Context, bizTradeNo string) (domain.Payment, error)
	HandleCallback(ctx context.Context, transaction *payments.Transaction) error
}
