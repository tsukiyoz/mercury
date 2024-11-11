package domain

import "github.com/wechatpay-apiv3/wechatpay-go/services/payments"

type Amount struct {
	Currency string
	Total    int64
}

type Payment struct {
	Amt         Amount
	BizTradeNo  string
	Description string
	Status      PaymentStatus
	TxnID       string
}

type PaymentStatus uint8

func (s PaymentStatus) AsUint8() uint8 {
	return uint8(s)
}

const (
	PaymentStatusUnknown PaymentStatus = iota
	PaymentStatusInit                  // 初始化
	PaymentStatusSuccess               // 支付成功
	PaymentStatusFailed                // 支付失败
	PaymentStatusRefund                // 退款
)

type Transaction payments.Transaction
