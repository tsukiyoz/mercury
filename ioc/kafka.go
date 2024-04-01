package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	events2 "github.com/tsukaychan/webook/interactive/events"
	"github.com/tsukaychan/webook/internal/events"
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

func NewConsumers(consumer *events2.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{consumer}
}
