package events

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaProducer(client sarama.Client) (*SaramaProducer, error) {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return &SaramaProducer{producer: p}, nil
}

func (sp *SaramaProducer) ProducePaymentEvent(ctx context.Context, event PaymentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, _, err = sp.producer.SendMessage(&sarama.ProducerMessage{
		Topic: event.Topic(),
		Key:   sarama.StringEncoder(event.BizTradeNo),
		Value: sarama.ByteEncoder(data),
	})
	return err
}
