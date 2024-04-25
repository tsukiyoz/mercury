package ioc

import (
	"fmt"
	"time"

	"github.com/tsukaychan/mercury/user/repository/dao"

	"github.com/spf13/viper"
	interactiveDao "github.com/tsukaychan/mercury/interactive/repository/dao"
	"github.com/tsukaychan/mercury/pkg/gormx/callbacks/metrics"
	"github.com/tsukaychan/mercury/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	gormPrometheus "gorm.io/plugin/prometheus"
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
		DBName:          "mercury",
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
		"mercury",
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
			tracing.WithDBName("mercury"),
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
