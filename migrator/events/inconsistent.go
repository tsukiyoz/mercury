package events

import (
	"context"
	"github.com/tsukaychan/mercury/migrator/validator"
)

type InconsistentEvent struct {
	ID        int64
	Direction validator.DirectionType
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
