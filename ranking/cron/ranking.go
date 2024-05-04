package cron

import (
	"context"
	"sync"
	"time"

	"github.com/tsukaychan/mercury/pkg/cronx"

	"github.com/tsukaychan/mercury/ranking/service"

	"github.com/tsukaychan/mercury/pkg/logger"

	rlock "github.com/gotomicro/redis-lock"
)

var _ cronx.Task = (*RankingJob)(nil)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration
	client  *rlock.Client
	key     string
	l       logger.Logger

	lock *rlock.Lock
	mu   sync.Mutex
}

func NewRankingJob(
	svc service.RankingService,
	timeout time.Duration,
	client *rlock.Client,
	l logger.Logger,
) *RankingJob {
	return &RankingJob{
		svc:     svc,
		timeout: timeout,
		client:  client,
		key:     "rlock:cron_job:ranking",
		l:       l,
	}
}

func (job *RankingJob) Name() string {
	return "ranking"
}

func (job *RankingJob) Run() error {
	job.mu.Lock()
	defer job.mu.Unlock()
	if job.lock == nil {
		// get distributed lock
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lock, err := job.client.Lock(ctx, job.key, job.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      3,
		}, time.Second)
		if err != nil {
			// distributed lock are held by other instance
			return nil
		}

		job.lock = lock
		go func() {
			// automatic renewal
			if err := lock.AutoRefresh(job.timeout/2, time.Second); err != nil {
				// renewal failed, strive to grab the lock the next time
				job.l.Error("renewal distributed lock failed", logger.Error(err))
				job.mu.Lock()
				job.lock = nil
				job.mu.Unlock()
			}
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), job.timeout)
	defer cancel()
	return job.svc.RankTopN(ctx)
}

func (job *RankingJob) Close() error {
	job.mu.Lock()
	lock := job.lock
	job.lock = nil
	job.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
