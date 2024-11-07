package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"github.com/lazywoo/mercury/internal/interactive/repository/dao"
	"github.com/lazywoo/mercury/pkg/ginx"
	"github.com/lazywoo/mercury/pkg/gormx/connpool"
	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/lazywoo/mercury/pkg/migrator/events"
	"github.com/lazywoo/mercury/pkg/migrator/events/fixer"
	"github.com/lazywoo/mercury/pkg/migrator/scheduler"
)

const topic = "migrator_interactives"

func InitFixDataConsumer(src SrcDB, dst DstDB, client sarama.Client, l logger.Logger) *fixer.Consumer[dao.Interactive] {
	consumer, err := fixer.NewConsumer[dao.Interactive](client, src, dst, topic, l)
	if err != nil {
		panic(err)
	}
	return consumer
}

func InitMigratorProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(p, topic)
}

func InitMigratorWeb(
	src SrcDB,
	dst DstDB,
	pool *connpool.DualWritePool,
	producer events.Producer,
	l logger.Logger,
) *ginx.Server {
	web := gin.Default()
	ginx.InitCounterVec(prometheus.CounterOpts{
		Namespace: "tsukiyo",
		Subsystem: "webook_interactive",
		Name:      "http_biz_code",
		Help:      "HTTP Request in GIN",
		ConstLabels: prometheus.Labels{
			"instance_id": "instance-0",
		},
	})
	s := scheduler.NewScheduler[dao.Interactive](src, dst, pool, producer, l)
	s.RegisterRoutes(web.Group("/migrator"))
	addr := viper.GetString("migrator.http.addr")
	return &ginx.Server{
		Engine: web,
		Addr:   addr,
	}
}

func InitDualWritePool(src SrcDB, dst DstDB) *connpool.DualWritePool {
	return connpool.NewDualWritePool(src, dst)
}
