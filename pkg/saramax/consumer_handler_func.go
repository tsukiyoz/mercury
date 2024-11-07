package saramax

import (
	"encoding/json"

	"github.com/IBM/sarama"

	"github.com/lazywoo/mercury/pkg/logger"
)

type Handler[T any] struct {
	l  logger.Logger
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](l logger.Logger, fn func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	return &Handler[T]{
		l:  l,
		fn: fn,
	}
}

func (hdl *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (hdl *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (hdl *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgCh := claim.Messages()
	for msg := range msgCh {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			hdl.l.Error("unmarshal message failed",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset),
			)
		}
		for i := 0; i < 3; i++ {
			err = hdl.fn(msg, t)
			if err == nil {
				break
			}
			if err != nil {
				hdl.l.Error("handle message failed",
					logger.String("topic", msg.Topic),
					logger.Int32("partition", msg.Partition),
					logger.Int64("offset", msg.Offset),
					logger.Error(err))
			}
		}
		if err != nil {
			hdl.l.Error("handle message retry reaching maximum retry limit",
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset),
				logger.Error(err))
		} else {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}
