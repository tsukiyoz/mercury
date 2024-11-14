package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/lazywoo/mercury/internal/payment/domain"
	"github.com/lazywoo/mercury/internal/payment/service/wechat"
	paymentv1 "github.com/lazywoo/mercury/pkg/api/payment/v1"
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

func (w *WechatServiceServer) NativePrePay(ctx context.Context, payment *paymentv1.PrePayRequest) (*paymentv1.NativePrePayResponse, error) {
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
