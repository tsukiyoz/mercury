package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusBuilder struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	InstanceID string
}

func (bdr *PrometheusBuilder) Build() gin.HandlerFunc {
	labels := []string{"method", "pattern", "status"}
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: bdr.Namespace,
		Subsystem: bdr.Subsystem,
		Name:      bdr.Name + "_response_time",
		Help:      bdr.Help,
		ConstLabels: prometheus.Labels{
			"instance_id": bdr.InstanceID,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(summary)

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: bdr.Namespace,
		Subsystem: bdr.Subsystem,
		Name:      bdr.Name + "_active_request",
		Help:      bdr.Help,
		ConstLabels: prometheus.Labels{
			"instance_id": bdr.InstanceID,
		},
	})
	prometheus.MustRegister(gauge)

	return func(ctx *gin.Context) {
		start := time.Now()
		gauge.Inc()
		defer func() {
			duration := time.Since(start)
			gauge.Dec()
			pattern := ctx.FullPath()
			if pattern == "" {
				pattern = "unknown"
			}
			summary.WithLabelValues(
				ctx.Request.Method,
				pattern,
				strconv.Itoa(ctx.Writer.Status()),
			).Observe(float64(duration.Milliseconds()))
		}()
		ctx.Next()
	}
}
