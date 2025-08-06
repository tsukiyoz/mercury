package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/tsukiyo/mercury/internal/sms/service"
)

type FailoverSMSService struct {
	svcs   []service.Service
	currId uint64
}

//func (f *FailoverSMSService) Send(ctx context.Context, tpl string, args []sms.ArgVal, phones ...string) error {
//	for _, svc := range f.svcs {
//		err := svc.Send(ctx, tpl, args, phones...)
//		// send success
//		if err == nil {
//			return nil
//		}
//		// log info and watch
//		log.Println(err)
//
//	}
//
//	return errors.New("all sms service send failed")
//}

func (f *FailoverSMSService) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	n := len(f.svcs)
	for t := 0; t < n; t++ {
		idx := atomic.AddUint64(&f.currId, 1)
		svc := f.svcs[idx]
		err := svc.Send(ctx, tpl, target, args, values)
		atomic.StoreUint64(&f.currId, idx)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			return err
		}
		// log info and watch
	}
	return errors.New("all sms service send failed")
}

func NewFailoverSMSService(svcs []service.Service) service.Service {
	return &FailoverSMSService{
		svcs:   svcs,
		currId: 0,
	}
}
