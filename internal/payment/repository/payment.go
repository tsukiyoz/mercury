package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lazywoo/mercury/internal/payment/domain"
	"github.com/lazywoo/mercury/internal/payment/repository/dao"
)

func NewPaymentRepository(dao dao.PaymentDAO) PaymentRepository {
	return &paymentRepository{
		dao: dao,
	}
}

type paymentRepository struct {
	dao dao.PaymentDAO
}

func (p *paymentRepository) AddPayment(ctx context.Context, payment domain.Payment) error {
	return p.dao.Insert(ctx, p.toEntity(payment))
}

func (p *paymentRepository) UpdatePayment(ctx context.Context, payment domain.Payment) error {
	return p.dao.UpdateTxnIDAndStatus(ctx, payment.BizTradeNo, payment.TxnID, payment.Status)
}

// FindExpiredPayments implements PaymentRepository.
func (p *paymentRepository) FindExpiredPayments(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error) {
	payments, err := p.dao.FindExpiredPayments(ctx, offset, limit, t)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Payment, 0, len(payments))
	for _, payment := range payments {
		res = append(res, p.toDomain(payment))
	}
	return res, nil
}

// GetPayment implements PaymentRepository.
func (p *paymentRepository) GetPayment(ctx context.Context, bizTradeNo string) (domain.Payment, error) {
	payment, err := p.dao.GetPayment(ctx, bizTradeNo)
	if err != nil {
		return domain.Payment{}, err
	}
	return p.toDomain(payment), nil
}

func (p *paymentRepository) toEntity(payment domain.Payment) dao.Payment {
	return dao.Payment{
		Amount:      payment.Amt.Total,
		Currency:    payment.Amt.Currency,
		Description: payment.Description,
		BizTradeNo:  payment.BizTradeNo,
		TxnID:       sql.NullString{String: payment.TxnID},
		Status:      uint8(payment.Status),
		Utime:       0,
		Ctime:       0,
	}
}

func (p *paymentRepository) toDomain(payment dao.Payment) domain.Payment {
	return domain.Payment{
		Amt:         domain.Amount{Currency: payment.Currency, Total: payment.Amount},
		BizTradeNo:  payment.BizTradeNo,
		Description: payment.Description,
		Status:      domain.PaymentStatus(payment.Status),
		TxnID:       payment.TxnID.String,
	}
}
