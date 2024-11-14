package cronjob

import (
	"context"
	"sync"
	"time"

	rlock "github.com/gotomicro/redis-lock"

	"github.com/lazywoo/mercury/internal/payment/service/wechat"
	"github.com/lazywoo/mercury/pkg/logger"
)

type SyncWechatOrderJob struct {
	svc     *wechat.NativePaymentService
	timeout time.Duration
	client  *rlock.Client
	l       logger.Logger
	key     string

	lock *rlock.Lock
	mu   sync.Mutex
}

func NewSyncWechatOrderJob(svc *wechat.NativePaymentService,
	timeout time.Duration,
	client *rlock.Client,
	l logger.Logger,
) *SyncWechatOrderJob {
	return &SyncWechatOrderJob{
		svc:     svc,
		timeout: timeout,
		client:  client,
		key:     "rlock:cron_job:sync_wechat_order",
		l:       l,
	}
}

func (s *SyncWechatOrderJob) Name() string {
	return "sync_wechat_order_job"
}

func (s *SyncWechatOrderJob) Run() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.l.Info("start sync wechat order job", logger.String("execute_at", time.Now().Format(time.DateTime)))
	if s.lock == nil {
		s.l.Info("try to get distributed lock", logger.String("key", s.key), logger.String("execute_at", time.Now().Format(time.DateTime)))
		// get distributed lock
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lock, err := s.client.Lock(ctx, s.key, s.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      3,
		}, time.Second)
		if err != nil {
			// distributed lock are held by other instance
			return nil
		}

		s.lock = lock
		go func() {
			// automatic renewal
			if err := lock.AutoRefresh(s.timeout/2, time.Second); err != nil {
				// renewal failed, strive to grab the lock the next time
				s.l.Error("renewal distributed lock failed", logger.Error(err))
				s.mu.Lock()
				s.lock = nil
				s.mu.Unlock()
			}
		}()
	}

	offset := 0
	const limit = 100
	now := time.Now().Add(-time.Minute * 3)
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
