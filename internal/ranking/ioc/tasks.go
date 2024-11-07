package ioc

import (
	"context"
	"log"
	"time"

	"github.com/lazywoo/mercury/pkg/cronx"

	cron2 "github.com/lazywoo/mercury/internal/ranking/cron"

	"github.com/lazywoo/mercury/internal/crontask/domain"
	"github.com/lazywoo/mercury/internal/crontask/service"

	service2 "github.com/lazywoo/mercury/internal/ranking/service"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"

	"github.com/lazywoo/mercury/pkg/logger"
)

// -------------------------------------------
// distributed task schedule base on redis
// -------------------------------------------

func InitRankingJob(svc service2.RankingService, rlockClient *rlock.Client, l logger.Logger) *cron2.RankingJob {
	return cron2.NewRankingJob(svc, time.Second*30, rlockClient, l)
}

func InitTasks(l logger.Logger, ranking *cron2.RankingJob) *cron.Cron {
	croj := cron.New(cron.WithSeconds())
	bdr := cronx.NewCronJobBuilder(prometheus.SummaryOpts{
		Namespace: "tsukiyo",
		Subsystem: "mercury",
		Name:      "cron_job",
		Help:      "metrics cron job",
	}, l)
	// @every 3m
	_, err := croj.AddJob("0 */3 * * * ?", bdr.Build(ranking))
	if err != nil {
		panic(err)
	}
	_, err = croj.AddJob("0 */1 * * * ?", &DummyJob{})
	if err != nil {
		panic(err)
	}
	return croj
}

type DummyJob struct{}

func (d DummyJob) Run() {
	log.Println("test cron job")
}

// -------------------------------------------
// distributed task schedule base on mysql
// -------------------------------------------

func InitScheduler(svc service.TaskService,
	executor cronx.Executor,
	l logger.Logger,
) *cronx.Scheduler {
	scheduler := cronx.NewScheduler(svc, l)
	scheduler.RegisterExecutor(executor)
	return scheduler
}

func InitLocalFuncExecutor(svc service2.RankingService) cronx.Executor {
	executor := cronx.NewLocalFuncExecutor()
	executor.AddLocalFunc("ranking", func(ctx context.Context, tsk domain.Task) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		return svc.RankTopN(ctx)
	})
	return executor
}
