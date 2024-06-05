package cronx

import (
	"strconv"
	"time"

	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
)

type CronJobBuilder struct {
	prom *prometheus.SummaryVec
	l    logger.Logger
}

func NewCronJobBuilder(opts prometheus.SummaryOpts, l logger.Logger) *CronJobBuilder {
	prom := prometheus.NewSummaryVec(opts, []string{"name", "success"})
	prometheus.MustRegister(prom)

	return &CronJobBuilder{
		l:    l,
		prom: prom,
	}
}

func (bdr *CronJobBuilder) Build(job Task) cron.Job {
	name := job.Name()
	return cronJobFuncAdapter(func() error {
		start := time.Now()
		bdr.l.Debug("cron job start",
			logger.String("name", name),
			logger.String("time", start.String()),
		)

		var success bool
		defer func() {
			bdr.l.Debug("cron job finish",
				logger.String("name", name),
				logger.String("time", start.String()),
			)
			duration := time.Since(start).Milliseconds()
			bdr.prom.WithLabelValues(name, strconv.FormatBool(success)).Observe(float64(duration))
		}()

		err := job.Run()
		success = err == nil
		if err != nil {
			bdr.l.Error("execute cron job failed",
				logger.Error(err),
				logger.String("job", name),
			)
		}

		return nil
	})
}

var _ cron.Job = (*cronJobFuncAdapter)(nil)

type cronJobFuncAdapter func() error

func (fn cronJobFuncAdapter) Run() {
	fn()
}
