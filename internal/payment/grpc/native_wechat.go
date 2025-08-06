package grpc

import (
	"context"

	"google.golang.org/grpc"

	paymentv1 "github.com/tsukiyo/mercury/api/gen/payment/v1"
	"github.com/tsukiyo/mercury/internal/payment/domain"
	"github.com/tsukiyo/mercury/internal/payment/service/wechat"
)

type WechatServiceServer struct {
	paymentv1.UnimplementedWechatPaymentServiceServer
	svc *wechat.NativePaymentService
}

func NewWechatPaymentServiceServer(svc *wechat.NativePaymentService) *WechatServiceServer {
	return &WechatServiceServer{
		svc: svc,
	}
}

func (a *WechatServiceServer) Register(server grpc.ServiceRegistrar) {
	paymentv1.RegisterWechatPaymentServiceServer(server, a)
}

func (w *WechatServiceServer) NativePrePay(ctx context.Context, payment *paymentv1.NativePrePayRequest) (*paymentv1.NativePrePayResponse, error) {
	codeUrl, err := w.svc.Prepay(ctx, domain.Payment{
		Amount:      domain.Amount{Currency: payment.Amount.Currency, Total: payment.Amount.Total},
		BizTradeNo:  payment.BizTradeNo,
		Description: payment.Description,
		Status:      domain.PaymentStatusInit,
	})
	if err != nil {
		return nil, err
	}
	return &paymentv1.NativePrePayResponse{
		CodeUrl: codeUrl,
	}, nil
}

func (w *WechatServiceServer) GetPayment(ctx context.Context, payment *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	res, err := w.svc.GetPayment(ctx, payment.BizTradeNo)
	if err != nil {
		return nil, err
	}
	return &paymentv1.GetPaymentResponse{
		Status: paymentv1.PaymentStatus(res.Status),
	}, nil
}

func (w *WechatServiceServer) RefundPayment(ctx context.Context, req *paymentv1.RefundPaymentRequest) (*paymentv1.RefundPaymentResponse, error) {
	err := w.svc.Refund(ctx, domain.Payment{BizTradeNo: req.BizTradeNo}, req.RefundReason)
	return &paymentv1.RefundPaymentResponse{}, err
}
