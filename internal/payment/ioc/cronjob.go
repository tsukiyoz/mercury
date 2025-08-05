package ioc

import (
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"

	"github.com/tsukiyo/mercury/internal/payment/cronjob"
	"github.com/tsukiyo/mercury/internal/payment/service/wechat"
	"github.com/tsukiyo/mercury/pkg/cronx"
	"github.com/tsukiyo/mercury/pkg/logger"
)

func InitCronJobs(l logger.Logger, syncWechatPaymentJob *cronjob.SyncWechatOrderJob) *cron.Cron {
	cronJob := cron.New(cron.WithSeconds())
	bdr := cronx.NewCronJobBuilder(prometheus.SummaryOpts{
		Namespace: "lazywoo",
		Subsystem: "mercury",
		Name:      "cron_job",
		Help:      "metrics cron job",
	}, l)
	// @every 3m
	_, err := cronJob.AddJob("0 */3 * * * ?", bdr.Build(syncWechatPaymentJob))
	if err != nil {
		panic(err)
	}
	return cronJob
}

func InitSyncWechatPaymentJob(svc *wechat.NativePaymentService,
	client *rlock.Client,
	l logger.Logger,
) *cronjob.SyncWechatOrderJob {
	return cronjob.NewSyncWechatOrderJob(svc, time.Second*3, client, l)
}
