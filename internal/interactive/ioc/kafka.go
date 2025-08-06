package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	"github.com/tsukiyo/mercury/internal/interactive/events"
	"github.com/tsukiyo/mercury/internal/interactive/repository/dao"
	migratorEvt "github.com/tsukiyo/mercury/pkg/migrator/events/fixer"
	"github.com/tsukiyo/mercury/pkg/saramax"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true

	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	syncProducer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return syncProducer
}

func NewConsumers(consumer *events.InteractiveReadEventConsumer, fix *migratorEvt.Consumer[dao.Interactive]) []saramax.Consumer {
	return []saramax.Consumer{consumer, fix}
}
