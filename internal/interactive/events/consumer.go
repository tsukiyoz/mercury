package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"github.com/tsukiyo/mercury/internal/interactive/repository"

	"github.com/tsukiyo/mercury/pkg/logger"
	"github.com/tsukiyo/mercury/pkg/saramax"
)

const topicReadEvent = "article_read_event"

type ReadEvent struct {
	Aid int64
	Uid int64
}

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
			// saramax.NewHandler[ReadEvent](consumer.l, consumer.Consume),
			saramax.NewBatchHandler[ReadEvent](consumer.l, consumer.BatchConsume),
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

func (consumer *InteractiveReadEventConsumer) StartBatch() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", consumer.client)
	if err != nil {
		return err
	}

	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicReadEvent},
			saramax.NewBatchHandler[ReadEvent](consumer.l, consumer.BatchConsume),
		)
		if err != nil {
			consumer.l.Error("exited consumption cycle exception", logger.Error(err))
		}
	}()

	return err
}

func (consumer *InteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage,
	evts []ReadEvent,
) error {
	ids := make([]int64, 0, len(msgs))
	for _, evt := range evts {
		ids = append(ids, evt.Aid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := consumer.repo.BatchIncrReadCnt(ctx, "article", ids)
	if err != nil {
		consumer.l.Error("batch increase read count failed",
			logger.Int64Slice("ids", ids),
			logger.Error(err),
		)
	}
	return nil
}
