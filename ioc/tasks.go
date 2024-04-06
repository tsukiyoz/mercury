package ioc

import (
	"context"
	"log"
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"github.com/tsukaychan/mercury/internal/domain"
	"github.com/tsukaychan/mercury/internal/service"
	"github.com/tsukaychan/mercury/internal/task"
	"github.com/tsukaychan/mercury/pkg/logger"
)

// -------------------------------------------
// distributed task schedule base on redis
// -------------------------------------------

func InitRankingJob(svc service.RankingService, rlockClient *rlock.Client, l logger.Logger) *task.RankingJob {
	return task.NewRankingJob(svc, time.Second*30, rlockClient, l)
}

func InitTasks(l logger.Logger, ranking *task.RankingJob) *cron.Cron {
	croj := cron.New(cron.WithSeconds())
	bdr := task.NewCronJobBuilder(prometheus.SummaryOpts{
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
	executor task.Executor,
	l logger.Logger,
) *task.Scheduler {
	scheduler := task.NewScheduler(svc, l)
	scheduler.RegisterExecutor(executor)
	return scheduler
}

func InitLocalFuncExecutor(svc service.RankingService) task.Executor {
	executor := task.NewLocalFuncExecutor()
	executor.AddLocalFunc("ranking", func(ctx context.Context, tsk domain.Task) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		return svc.RankTopN(ctx)
	})
	return executor
}
