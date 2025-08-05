package saramax

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"

	"github.com/tsukiyo/mercury/pkg/logger"
)

type BatchHandler[Evt any] struct {
	l             logger.Logger
	fn            func(msg []*sarama.ConsumerMessage, ts []Evt) error
	batchSize     int
	batchDuration time.Duration
}

type Option[Evt any] func(handler *BatchHandler[Evt])

func NewBatchHandler[Evt any](l logger.Logger, fn func(msg []*sarama.ConsumerMessage, ts []Evt) error, opts ...Option[Evt]) *BatchHandler[Evt] {
	hdl := &BatchHandler[Evt]{l: l, fn: fn, batchSize: 10, batchDuration: time.Second}
	for _, opt := range opts {
		opt(hdl)
	}
	return hdl
}

func (hdl *BatchHandler[Evt]) WithBatchSize(batchSize int) Option[Evt] {
	return func(hdl *BatchHandler[Evt]) {
		hdl.batchSize = batchSize
	}
}

func (hdl *BatchHandler[Evt]) WithBatchDuration(batchDuration time.Duration) Option[Evt] {
	return func(hdl *BatchHandler[Evt]) {
		hdl.batchDuration = batchDuration
	}
}

func (hdl *BatchHandler[Evt]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (hdl *BatchHandler[Evt]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (hdl *BatchHandler[Evt]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgCh := claim.Messages()
	batchSize, batchDuration := hdl.batchSize, hdl.batchDuration
	for {
		ctx, cancel := context.WithTimeout(context.Background(), batchDuration)
		done := false
		var last *sarama.ConsumerMessage
		msgs := make([]*sarama.ConsumerMessage, 0, batchSize)
		evts := make([]Evt, 0, batchSize)
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgCh:
				if !ok {
					cancel()
					return nil
				}
				last = msg
				var t Evt
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					hdl.l.Error("unmarshal message failed",
						logger.String("topic", msg.Topic),
						logger.Int32("partition", msg.Partition),
						logger.Int64("offset", msg.Offset),
						logger.Error(err),
					)
					session.MarkMessage(msg, "")
					continue
				}
				msgs = append(msgs, msg)
				evts = append(evts, t)
			}
		}
		cancel()
		if len(msgs) == 0 {
			continue
		}
		err := hdl.fn(msgs, evts)
		if err != nil {
			hdl.l.Error("call business batch interface failed",
				logger.Error(err),
			)
		}
		if last != nil {
			session.MarkMessage(last, "")
		}
	}
}
