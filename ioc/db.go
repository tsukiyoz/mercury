package ioc

import (
	"fmt"
	"time"

	"github.com/tsukaychan/webook/pkg/gormx/callbacks/prometheus"

	"github.com/tsukaychan/webook/internal/repository/dao"
	"github.com/tsukaychan/webook/pkg/logger"
	gormPrometheus "gorm.io/plugin/prometheus"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
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

	prom := prometheus.NewCallbacks(
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

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g("[SQL]", logger.Field{Key: "args", Value: fmt.Sprintf(msg, args...)})
}
