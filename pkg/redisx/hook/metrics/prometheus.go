package metrics

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type PrometheusHook struct {
	summaryVec *prometheus.SummaryVec
}

func NewPrometheusHook(
	namespace string,
	subsystem string,
	instanceId string,
	name string,
) *PrometheusHook {
	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		ConstLabels: prometheus.Labels{
			"instance_id": instanceId,
		},
	}, []string{"cmd", "hit_cache"})
	prometheus.MustRegister(summaryVec)
	return &PrometheusHook{
		summaryVec: summaryVec,
	}
}

func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		var err error
		defer func() {
			keyExist := err == redis.Nil
			p.summaryVec.WithLabelValues(cmd.Name(),
				strconv.FormatBool(keyExist)).
				Observe(float64(time.Since(start)))
		}()
		err = next(ctx, cmd)
		return err
	}
}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
