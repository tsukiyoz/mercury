package metrics

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type Callbacks struct {
	Namespace  string
	Subsystem  string
	Name       string
	InstanceID string
	Help       string
	summaryVec *prometheus.SummaryVec
}

func NewCallbacks(namespace string, subsystem string, name string, instanceID string, help string) *Callbacks {
	return &Callbacks{
		Namespace:  namespace,
		Subsystem:  subsystem,
		Name:       name,
		InstanceID: instanceID,
		Help:       help,
	}
}

func (callbacks *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (callbacks *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			log.Printf("callbacks get start time failed!")
			return
		}
		duration := time.Since(startTime).Milliseconds()
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		callbacks.summaryVec.WithLabelValues(typ, table).Observe(float64(duration))
	}
}

func (callbacks *Callbacks) Register(db *gorm.DB) error {
	summaryVec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: callbacks.Namespace,
			Subsystem: callbacks.Subsystem,
			Name:      callbacks.Name,
			Help:      callbacks.Help,
			ConstLabels: prometheus.Labels{
				"db_name":     db.Name(),
				"instance_id": callbacks.InstanceID,
			},
			Objectives: map[float64]float64{
				0.5:  0.01,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"type", "table"})
	prometheus.MustRegister(summaryVec)
	callbacks.summaryVec = summaryVec

	err := db.Callback().Create().Before("*").
		Register("prometheus_create_before", callbacks.before())
	if err != nil {
		return err
	}

	err = db.Callback().Create().After("*").
		Register("prometheus_create_after", callbacks.after("create"))
	if err != nil {
		return err
	}

	err = db.Callback().Update().Before("*").
		Register("prometheus_update_before", callbacks.before())
	if err != nil {
		return err
	}

	err = db.Callback().Update().After("*").
		Register("prometheus_update_after", callbacks.after("update"))
	if err != nil {
		return err
	}

	err = db.Callback().Delete().Before("*").
		Register("prometheus_delete_before", callbacks.before())
	if err != nil {
		return err
	}

	err = db.Callback().Delete().After("*").
		Register("prometheus_delete_after", callbacks.after("delete"))
	if err != nil {
		return err
	}

	err = db.Callback().Raw().Before("*").
		Register("prometheus_raw_before", callbacks.before())
	if err != nil {
		return err
	}

	err = db.Callback().Raw().After("*").
		Register("prometheus_raw_after", callbacks.after("raw"))
	if err != nil {
		return err
	}

	err = db.Callback().Row().Before("*").
		Register("prometheus_row_before", callbacks.before())
	if err != nil {
		return err
	}

	err = db.Callback().Row().After("*").
		Register("prometheus_row_after", callbacks.after("row"))
	if err != nil {
		return err
	}

	err = db.Callback().Query().Before("*").
		Register("prometheus_query_before", callbacks.before())
	if err != nil {
		return err
	}

	err = db.Callback().Query().After("*").
		Register("prometheus_query_after", callbacks.after("query"))
	if err != nil {
		return err
	}

	return nil
}
