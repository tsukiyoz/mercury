package ioc

import (
	"fmt"
	"time"

	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	gormPrometheus "gorm.io/plugin/prometheus"

	interactiveDao "github.com/tsukaychan/webook/interactive/repository/dao"
	"github.com/tsukaychan/webook/internal/repository/dao"
	"github.com/tsukaychan/webook/pkg/gormx/callbacks/metrics"
	"github.com/tsukaychan/webook/pkg/logger"
)

func InitDB(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}

	var cfg Config
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: gormLogger.New(gormLoggerFunc(l.Debug), gormLogger.Config{
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  gormLogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}

	// metrics
	err = db.Use(gormPrometheus.New(gormPrometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		MetricsCollector: []gormPrometheus.MetricsCollector{
			&gormPrometheus.MySQL{
				VariableNames: []string{"threads_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	prom := metrics.NewCallbacks(
		"tsukiyo",
		"webook",
		"prometheus_query",
		"instance-0",
		"metrics gorm db query",
	)
	err = prom.Register(db)
	if err != nil {
		panic(err)
	}

	// tracing
	db.Use(
		tracing.NewPlugin(
			tracing.WithDBName("webook"),
			tracing.WithQueryFormatter(func(query string) string {
				l.Debug("query", logger.String("query", query))
				return query
			}),
			tracing.WithoutMetrics(),
			tracing.WithoutQueryVariables(),
		),
	)

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	err = interactiveDao.InitTable(db)
	if err != nil {
		panic(err)
	}

	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g("[SQL]", logger.Field{Key: "args", Value: fmt.Sprintf(msg, args...)})
}
