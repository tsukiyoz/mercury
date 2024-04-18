package fixer

import (
	"github.com/IBM/sarama"
	"github.com/tsukaychan/mercury/pkg/logger"
	"github.com/tsukaychan/mercury/pkg/migrator"
	"github.com/tsukaychan/mercury/pkg/migrator/fixer"
	"gorm.io/gorm"
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
