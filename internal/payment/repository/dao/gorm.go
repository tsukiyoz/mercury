package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/tsukiyo/mercury/internal/payment/domain"
)

type GORMPaymentDAO struct {
	db *gorm.DB
}

func NewGORMPaymentDAO(db *gorm.DB) PaymentDAO {
	return &GORMPaymentDAO{
		db: db,
	}
}

func (p *GORMPaymentDAO) Insert(ctx context.Context, payment Payment) error {
	now := time.Now().UnixMilli()
	payment.Utime = now
	payment.Ctime = now
	return p.db.WithContext(ctx).Create(&payment).Error
}

func (p *GORMPaymentDAO) GetPayment(ctx context.Context, bizTradeNo string) (Payment, error) {
	var res Payment
	err := p.db.WithContext(ctx).Where("biz_trade_no = ?", bizTradeNo).First(&res).Error
	return res, err
}

func (p *GORMPaymentDAO) FindExpiredPayments(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error) {
	var res []Payment
	err := p.db.WithContext(ctx).Where("status = ? AND utime < ?",
		domain.PaymentStatusInit.AsUint8(), t.UnixMilli()).
		Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (p *GORMPaymentDAO) UpdateTxnIDAndStatus(ctx context.Context,
	bizTradeNo string,
	txnID string,
	status domain.PaymentStatus,
) error {
	return p.db.WithContext(ctx).Model(&Payment{}).
		Where("biz_trade_no = ?", bizTradeNo).
		Updates(map[string]any{
			"txn_id": txnID,
			"status": status.AsUint8(),
			"utime":  time.Now().UnixMilli(),
		}).
		Error
}
