package migrator

type Entity interface {
	ID() int64
	Equal(v Entity) bool
}
