package circuitbreaker

import (
	"testing"

	"github.com/go-kratos/aegis/circuitbreaker/sre"
)

func TestCircuit(t *testing.T) {
	b := sre.NewBreaker()
	for i := 0; i < 1000; i++ {
		b.MarkSuccess()
	}
	for i := 0; i < 100; i++ {
		b.MarkFailed()
	}

	err := b.Allow()
	t.Logf("err=%v", err) // err=<nil>
}
