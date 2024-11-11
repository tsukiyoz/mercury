package events

type PaymentEvent struct {
	BizTradeNo string
	Status     uint8
}

func (PaymentEvent) Topic() string {
	return "payment_events"
}
