package wechat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"

	"github.com/lazywoo/mercury/internal/payment/domain"
	"github.com/lazywoo/mercury/internal/payment/events"
	"github.com/lazywoo/mercury/internal/payment/repository"
	"github.com/lazywoo/mercury/pkg/logger"
)

var errUnknownTransactionState = errors.New("unknwon wechat transaction status")

type NativePaymentService struct {
	svc                  *native.NativeApiService
	appID                string
	mchID                string
	notifyURL            string
	repo                 repository.PaymentRepository
	l                    logger.Logger
	producer             events.Producer
	callbackTypeToStatus map[string]domain.PaymentStatus
}

func NewNativePaymentService(
	svc *native.NativeApiService,
	appID string,
	mchID string,
	repo repository.PaymentRepository,
	l logger.Logger,
	producer events.Producer,
) *NativePaymentService {
	return &NativePaymentService{
		svc:       svc,
		appID:     appID,
		mchID:     mchID,
		notifyURL: "http://test.lazywoo.com/pay/callback",
		repo:      repo,
		l:         l,
		producer:  producer,
		callbackTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"CLOSED":   domain.PaymentStatusFailed,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
		},
	}
}

func (n *NativePaymentService) Prepay(ctx context.Context, payment domain.Payment) (string, error) {
	err := n.repo.AddPayment(ctx, payment)
	if err != nil {
		return "", err
	}
	resp, _, err := n.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(n.appID),
		Mchid:       core.String(n.mchID),
		Description: core.String(payment.Description),
		OutTradeNo:  core.String(payment.BizTradeNo),
		TimeExpire:  core.Time(time.Now().Add(time.Minute * 30)),
		NotifyUrl:   core.String(n.notifyURL),
		Amount: &native.Amount{
			Currency: core.String(payment.BizTradeNo),
			Total:    core.Int64(payment.Amt.Total),
		},
	})
	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil
}

func (n *NativePaymentService) SyncInfo(ctx context.Context, bizTradeNo string) error {
	txn, _, err := n.svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(bizTradeNo),
		Mchid:      core.String(n.mchID),
	})
	if err != nil {
		return err
	}
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	status, ok := n.callbackTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, %s", errUnknownTransactionState, *txn.TradeState)
	}
	payment := domain.Payment{
		BizTradeNo: *txn.OutTradeNo,
		TxnID:      *txn.TransactionId,
		Status:     status,
	}
	err := n.repo.UpdatePayment(ctx, payment)
	if err != nil {
		return err
	}
	err = n.producer.ProducePaymentEvent(ctx, events.PaymentEvent{
		BizTradeNo: payment.BizTradeNo,
		Status:     payment.Status.AsUint8(),
	})
	if err != nil {
		n.l.Error("send payment event failed", logger.Error(err),
			logger.String("biz_trade_no", payment.BizTradeNo))
		return err
	}
	return nil
}

func (n *NativePaymentService) FindExpiredPayments(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayments(ctx, offset, limit, t)
}

func (n *NativePaymentService) GetPayment(ctx context.Context, bizTradeNo string) (domain.Payment, error) {
	return n.repo.GetPayment(ctx, bizTradeNo)
}

func (n *NativePaymentService) HandleCallback(ctx context.Context, transaction *payments.Transaction) error {
	return n.updateByTxn(ctx, transaction)
}
