package events

type Producer interface {
	ProduceReadEvent(evt ReadEvent) error
}

type Consumer interface {
	Start() error
}
