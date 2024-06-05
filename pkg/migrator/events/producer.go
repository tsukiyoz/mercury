package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error
}

type SaramaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewSaramaProducer(p sarama.SyncProducer, topic string) *SaramaProducer {
	return &SaramaProducer{
		producer: p,
		topic:    topic,
	}
}

// ProduceInconsistentEvent produces an inconsistent event to the kafka topic
func (s *SaramaProducer) ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(data),
	})
	return err
}
