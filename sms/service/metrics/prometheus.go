package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/lazywoo/mercury/sms/service"
)

type PrometheusDecorator struct {
	svc        service.Service
	summaryVec *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc service.Service,
	namespace string,
	subsystem string,
	instanceId string,
	help string,
	name string,
) service.Service {
	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
		ConstLabels: prometheus.Labels{
			"instance_id": instanceId,
		},
		Objectives: map[float64]float64{
			0.9:   0.01,
			0.95:  0.001,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{
		"tpl",
	})
	prometheus.MustRegister(summaryVec)
	return &PrometheusDecorator{
		svc:        svc,
		summaryVec: summaryVec,
	}
}

func (p *PrometheusDecorator) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	startTime := time.Now()
	defer func() {
		p.summaryVec.WithLabelValues(tpl).Observe(float64(time.Since(startTime).Milliseconds()))
	}()
	return p.svc.Send(ctx, tpl, target, args, values)
}
