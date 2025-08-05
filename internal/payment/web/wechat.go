package web

import (
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"

	"github.com/tsukiyo/mercury/internal/payment/service/wechat"
	"github.com/tsukiyo/mercury/pkg/ginx"
	"github.com/tsukiyo/mercury/pkg/logger"
)

type WechatHandler struct {
	handler *notify.Handler
	l       logger.Logger
	svc     *wechat.NativePaymentService
}

func NewWechatHandler(hdl *notify.Handler, l logger.Logger, svc *wechat.NativePaymentService) *WechatHandler {
	return &WechatHandler{
		handler: hdl,
		l:       l,
		svc:     svc,
	}
}

func (h *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.POST("/pay/callback", ginx.Wrap(h.HandleNative))
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) (ginx.Result, error) {
	transaction := &payments.Transaction{}
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		return ginx.Result{}, err
	}
	err = h.svc.HandleCallback(ctx, transaction)
	return ginx.Result{}, err
}
