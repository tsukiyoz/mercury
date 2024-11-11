package job

import (
	"context"
	"time"

	"github.com/lazywoo/mercury/internal/payment/service/wechat"
	"github.com/lazywoo/mercury/pkg/logger"
)

type SyncWechatOrderJob struct {
	svc *wechat.NativePaymentService
	l   logger.Logger
}

func (s *SyncWechatOrderJob) Name() string {
	return "sync_wechat_order_job"
}

func (s *SyncWechatOrderJob) Run() error {
	offset := 0
	const limit = 100
	now := time.Now().Add(-time.Minute * 30)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		payments, err := s.svc.FindExpiredPayments(ctx, offset, limit, now)
		if err != nil {
			return err
		}
		if len(payments) == 0 {
			return nil
		}
		for _, payment := range payments {
			ictx, icancel := context.WithTimeout(context.Background(), time.Second)
			err = s.svc.SyncInfo(ictx, payment.BizTradeNo)
			if err != nil {
				s.l.Error("sync wechat payment info failed",
					logger.String("biz_trade_no", payment.BizTradeNo), logger.Error(err))
			}
			icancel()
		}
		cancel()
		offset += len(payments)
	}
}
