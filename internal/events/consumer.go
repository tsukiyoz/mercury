package events

import (
	"context"
	"time"

	"github.com/tsukaychan/webook/pkg/saramax"

	"github.com/IBM/sarama"
	"github.com/tsukaychan/webook/internal/repository"
	"github.com/tsukaychan/webook/pkg/logger"
)

var _ Consumer = (*InteractiveReadEventConsumer)(nil)

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewInteractiveReadEventConsumer(client sarama.Client,
	repo repository.InteractiveRepository,
	l logger.Logger,
) *InteractiveReadEventConsumer {
	intrCsr := &InteractiveReadEventConsumer{
		repo:   repo,
		client: client,
		l:      l,
	}
	return intrCsr
}

func (consumer *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", consumer.client)
	if err != nil {
		return err
	}

	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicReadEvent},
			saramax.NewHandler[ReadEvent](consumer.l, consumer.Consume),
		)
		if err != nil {
			consumer.l.Error("exited consumption cycle exception", logger.Error(err))
		}
	}()

	return err
}

func (consumer *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, evt ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return consumer.repo.IncrReadCnt(ctx, "article", evt.Aid)
}

func (consumer *InteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage,
	evts []ReadEvent,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	bizs := make([]string, 0, len(msgs))
	ids := make([]int64, 0, len(msgs))
	for _, evt := range evts {
		bizs = append(bizs, "article")
		ids = append(ids, evt.Uid)
	}
	return consumer.repo.BatchIncrReadCnt(ctx, bizs, ids)
}
