package dao

import (
	"context"
	"database/sql"
	"time"

	"github.com/tsukiyo/mercury/internal/payment/domain"
)

type PaymentDAO interface {
	Insert(ctx context.Context, payment Payment) error
	UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo string, txnID string, status domain.PaymentStatus) error
	FindExpiredPayments(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error)
	GetPayment(ctx context.Context, bizTradeNo string) (Payment, error)
}

type Payment struct {
	Id          int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Amount      int64
	Currency    string
	Description string         `gorm:"description"`
	BizTradeNo  string         `gorm:"column:biz_trade_no;type:varchar(256);unique"`
	TxnID       sql.NullString `gorm:"column:txn_id;type:varchar(128);unique"`
	Status      uint8
	Utime       int64
	Ctime       int64
}
