package fixer

import (
	"context"
	"errors"
	"time"

	"github.com/IBM/sarama"
	"gorm.io/gorm"

	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/lazywoo/mercury/pkg/migrator"
	"github.com/lazywoo/mercury/pkg/migrator/events"
	"github.com/lazywoo/mercury/pkg/migrator/fixer"
	"github.com/lazywoo/mercury/pkg/saramax"
)

type Consumer[T migrator.Entity] struct {
	client   sarama.Client
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
	topic    string
	l        logger.Logger
}

func NewConsumer[T migrator.Entity](client sarama.Client, src, dst *gorm.DB, topic string, l logger.Logger) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}

	dstFirst, err := fixer.NewOverrideFixer[T](dst, src)
	if err != nil {
		return nil, err
	}

	return &Consumer[T]{
		client:   client,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
		topic:    topic,
		l:        l,
	}, nil
}

func (r *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix", r.client)
	if err != nil {
		return err
	}

	go func() {
		err := cg.Consume(context.Background(), []string{r.topic}, saramax.NewHandler[events.InconsistentEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("exit consume loop error", logger.Error(err))
		}
	}()

	return err
}

func (r *Consumer[T]) Consume(msg *sarama.ConsumerMessage, evt events.InconsistentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	switch evt.Direction {
	case migrator.DirectionToTarget:
		return r.srcFirst.Fix(ctx, evt)
	case migrator.DirectionToBase:
		return r.dstFirst.Fix(ctx, evt)
	default:
		return errors.New("unknown direction")
	}
}
