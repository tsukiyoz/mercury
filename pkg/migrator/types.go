package migrator

type Entity interface {
	ID() int64
	Equal(v Entity) bool
}

type Direction uint8

const (
	DirectionToBase Direction = iota
	DirectionToTarget
)
