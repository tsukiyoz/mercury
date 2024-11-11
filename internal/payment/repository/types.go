package repository

import (
	"context"
	"time"

	"github.com/lazywoo/mercury/internal/payment/domain"
)

type PaymentRepository interface {
	AddPayment(ctx context.Context, payment domain.Payment) error
	UpdatePayment(ctx context.Context, payment domain.Payment) error
	FindExpiredPayments(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error)
	GetPayment(ctx context.Context, bizTradeNo string) (domain.Payment, error)
}
