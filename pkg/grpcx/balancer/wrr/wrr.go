package wrr

import (
	"context"
	"sync"

	"github.com/spf13/cast"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const name = "custom-wrr"

func init() {
	balancer.Register(base.NewBalancerBuilder(
		name,
		&PickerBuilder{},
		base.Config{
			HealthCheck: true,
		},
	))
}

type PickerBuilder struct{}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	for sc, scinfo := range info.ReadySCs {
		cc := &conn{cc: sc, available: true}

		cc.weight = cast.ToInt(scinfo.Address.Attributes.Value("weight"))

		if cc.weight == 0 {
			cc.weight = 10
		}

		conns = append(conns, cc)
	}
	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*conn
	mu    sync.Mutex
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var totalWeight int
	var target *conn

	p.mu.Lock()
	for _, cc := range p.conns {
		if !cc.available {
			continue
		}
		totalWeight += cc.weight
		cc.value += cc.weight
		if target == nil || cc.value > target.value {
			target = cc
		}
	}

	target.value -= totalWeight
	p.mu.Unlock()

	return balancer.PickResult{
		SubConn: target.cc,
		Done: func(info balancer.DoneInfo) {
			err := info.Err
			if err == nil {
				// dynamic increase weight
				return
			}

			// handle error
			switch err {
			case context.Canceled:
				return
			case context.DeadlineExceeded:
			// decrease weight
			default:
				sts, ok := status.FromError(err)
				if !ok {
					return
				}
				code := sts.Code()
				switch code {
				case codes.Unavailable:
					// fusing
					target.available = false
					go func() {
						// do health check
						if p.checkHealth(target) {
							// update available
							target.available = true
							// update value, make it smoothly
							// TODO with traffic control
						}
					}()
				case codes.ResourceExhausted:
					// limited
					// better low weight and value
					// to decrease the probability of being selected
				}
			}
		},
	}, nil
}

func (p *Picker) checkHealth(cc *conn) bool {
	return true
}

type conn struct {
	weight    int
	value     int
	cc        balancer.SubConn
	available bool
}
