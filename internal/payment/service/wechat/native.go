package wechat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"

	"github.com/tsukiyo/mercury/internal/payment/domain"
	"github.com/tsukiyo/mercury/internal/payment/events"
	"github.com/tsukiyo/mercury/internal/payment/repository"
	"github.com/tsukiyo/mercury/pkg/logger"
)

var errUnknownTransactionState = errors.New("unknwon wechat transaction status")

type NativePaymentService struct {
	nativeAPI            *native.NativeApiService
	refundAPI            *refunddomestic.RefundsApiService
	appID                string
	mchID                string
	notifyURL            string
	repo                 repository.PaymentRepository
	l                    logger.Logger
	producer             events.Producer
	callbackTypeToStatus map[string]domain.PaymentStatus
}

func NewNativePaymentService(
	natieAPI *native.NativeApiService,
	refundAPI *refunddomestic.RefundsApiService,
	appID string,
	mchID string,
	repo repository.PaymentRepository,
	l logger.Logger,
	producer events.Producer,
) *NativePaymentService {
	return &NativePaymentService{
		nativeAPI: natieAPI,
		refundAPI: refundAPI,
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
	resp, _, err := n.nativeAPI.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(n.appID),
		Mchid:       core.String(n.mchID),
		Description: core.String(payment.Description),
		OutTradeNo:  core.String(payment.BizTradeNo),
		TimeExpire:  core.Time(time.Now().Add(time.Minute * 30)),
		NotifyUrl:   core.String(n.notifyURL),
		Amount: &native.Amount{
			Currency: core.String(payment.Amount.Currency),
			Total:    core.Int64(payment.Amount.Total),
		},
	})
	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil
}

func (n *NativePaymentService) SyncInfo(ctx context.Context, bizTradeNo string) error {
	txn, _, err := n.nativeAPI.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
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
	payment, err := n.repo.GetPayment(ctx, bizTradeNo)
	if err != nil {
		return domain.Payment{}, err
	}
	if payment.Status == domain.PaymentStatusSuccess || payment.Status == domain.PaymentStatusRefund {
		return payment, nil
	}
	// 慢路径
	err = n.SyncInfo(ctx, bizTradeNo)
	if err != nil {
		return domain.Payment{}, err
	}
	return n.repo.GetPayment(ctx, bizTradeNo)
}

func (n *NativePaymentService) Refund(ctx context.Context, refund domain.Payment, reason string) error {
	oldPayment, err := n.repo.GetPayment(ctx, refund.BizTradeNo)
	if err != nil {
		return err
	}
	if oldPayment.Status != domain.PaymentStatusSuccess {
		return errors.New("not payment paid or payment is in refunding")
	}
	oldPayment.Status = domain.PaymentStatusInit
	err = n.repo.UpdatePayment(ctx, oldPayment)
	if err != nil {
		return err
	}

	_, result, err := n.refundAPI.Create(ctx, refunddomestic.CreateRequest{
		TransactionId: core.String(oldPayment.TxnID),
		OutTradeNo:    core.String(oldPayment.BizTradeNo),
		OutRefundNo:   core.String(oldPayment.BizTradeNo),
		Reason:        core.String(reason),
		NotifyUrl:     core.String(n.notifyURL),
		Amount: &refunddomestic.AmountReq{
			Currency: core.String(oldPayment.Amount.Currency),
			Refund:   core.Int64(oldPayment.Amount.Total),
			Total:    core.Int64(oldPayment.Amount.Total),
		},
	})
	if err != nil {
		bs, _ := io.ReadAll(result.Response.Body)
		var resultMap map[string]any
		_ = json.Unmarshal(bs, &resultMap)
		n.l.Error(
			"refund failed",
			logger.Error(err),
			logger.String("biz_trade_no", refund.BizTradeNo),
			logger.Int32("result.status_code", int32(result.Response.StatusCode)),
		)
		return fmt.Errorf("refund failed, %v", resultMap["message"])
	}
	return nil
}

func (n *NativePaymentService) HandleCallback(ctx context.Context, transaction *payments.Transaction) error {
	return n.updateByTxn(ctx, transaction)
}
