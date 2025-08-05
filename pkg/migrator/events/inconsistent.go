package events

import (
	"context"

	"github.com/tsukiyo/mercury/pkg/migrator"
)

type InconsistentEvent struct {
	ID        int64
	Direction migrator.Direction
	Type      InconsistentEventType
}

type InconsistentEventProducer interface {
	ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error
}

type InconsistentEventType uint

const (
	InconsistentEventTypeTargetMissing InconsistentEventType = iota
	InconsistentEventTypeBaseMissing
	InconsistentEventTypeNotEqual
)
