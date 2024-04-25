package async

import (
	"context"
	"time"

	"github.com/tsukaychan/mercury/internal/domain"
	"github.com/tsukaychan/mercury/internal/repository"
	"github.com/tsukaychan/mercury/pkg/logger"
	"github.com/tsukaychan/mercury/sms/service"
)

var _ service.Service = (*SMSService)(nil)

type SMSService struct {
	svc       service.Service
	asyncRepo repository.AsyncSmsRepository
	logger    logger.Logger
}

func NewSMSService(svc service.Service, repo repository.AsyncSmsRepository, logger logger.Logger) service.Service {
	s := &SMSService{
		svc:       svc,
		asyncRepo: repo,
		logger:    logger,
	}
	return s
}

func (s *SMSService) StartAsyncCycle() {
	for {
		s.AsyncSend()
	}
}

func (s *SMSService) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	if s.needAsync() {
		err := s.asyncRepo.Insert(ctx, domain.AsyncSms{
			TplId:  tpl,
			Target: target,
			Args:   args,
			Values: values,
		})
		return err
	}
	return s.Send(ctx, tpl, target, args, values)
}

func (s *SMSService) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	asyncSms, err := s.asyncRepo.PreemptWaitingSMS(ctx)
	cancel()

	switch err {
	case nil:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = s.svc.Send(ctx, asyncSms.TplId, asyncSms.Target, asyncSms.Args, asyncSms.Values)
		if err != nil {
			s.logger.Error("execute async send sms failed", logger.Error(err), logger.Int64("id", asyncSms.Id))
		}

		result := err == nil
		err = s.asyncRepo.ReportScheduleResult(ctx, asyncSms.Id, result)
		if err != nil {
			s.logger.Error("execute async send sms result, but report to repository failed", logger.Error(err),
				logger.Bool("result", result), logger.Int64("id", asyncSms.Id))
		}
	case repository.ErrWaitingSMSNotFound:
		time.Sleep(time.Second)
	default:
		s.logger.Error("preempt async sms task failed", logger.Error(err))
		time.Sleep(time.Second)
	}
}

func (s *SMSService) needAsync() bool {
	// threshold, change rate
	// error rate, response time
	// when quit async ?
	// 1. after 1 min
	// 2. release 1% of the traffic to attempt synchronization
	return true
}
