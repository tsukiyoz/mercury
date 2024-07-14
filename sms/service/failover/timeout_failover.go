package failover

import (
	"context"
	"sync/atomic"

	"github.com/lazywoo/mercury/sms/service"
)

type TimeoutFailoverSMSService struct {
	svcs       []service.Service
	lastUsedId uint32
	timeoutCnt uint32
	threshold  uint32
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	idx := atomic.LoadUint32(&t.lastUsedId)
	cnt := atomic.LoadUint32(&t.timeoutCnt)
	if cnt > t.threshold {
		// switch
		newIdx := (idx + 1) % uint32(len(t.svcs))
		if atomic.CompareAndSwapUint32(&t.lastUsedId, idx, newIdx) {
			atomic.StoreUint32(&t.timeoutCnt, 0)
		}
		idx = atomic.LoadUint32(&t.lastUsedId)
	}

	svc := t.svcs[idx]
	err := svc.Send(ctx, tpl, target, args, values)
	switch err {
	case context.DeadlineExceeded:
		atomic.AddUint32(&t.timeoutCnt, 1)
		return err
	case nil:
		atomic.StoreUint32(&t.timeoutCnt, 0)
		return nil
	default:
		return err
	}
}

func NewTimeoutFailoverSMSService() service.Service {
	return &TimeoutFailoverSMSService{}
}
