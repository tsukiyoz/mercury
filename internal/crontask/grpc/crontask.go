package grpc

import (
	"context"

	"github.com/lazywoo/mercury/internal/crontask/domain"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/lazywoo/mercury/internal/crontask/service"
	crontaskv1 "github.com/lazywoo/mercury/pkg/api/crontask/v1"
	"google.golang.org/grpc"
)

type CronJobServiceServer struct {
	crontaskv1.UnimplementedCronTaskServiceServer
	svc service.TaskService
}

func NewCronJobServiceServer(svc service.TaskService) *CronJobServiceServer {
	return &CronJobServiceServer{
		svc: svc,
	}
}

func (c *CronJobServiceServer) Register(server grpc.ServiceRegistrar) {
	crontaskv1.RegisterCronTaskServiceServer(server, c)
}

func (c *CronJobServiceServer) Preempt(ctx context.Context, req *crontaskv1.PreemptRequest) (*crontaskv1.PreemptResponse, error) {
	task, err := c.svc.Preempt(ctx)
	return &crontaskv1.PreemptResponse{
		Crontask: convertToV(task),
	}, err
}

func (c *CronJobServiceServer) ResetNextTime(ctx context.Context, req *crontaskv1.ResetNextTimeRequest) (*crontaskv1.ResetNextTimeResponse, error) {
	err := c.svc.ResetNextTime(ctx, convertToDomain(req.Task))
	return &crontaskv1.ResetNextTimeResponse{}, err
}

func (c *CronJobServiceServer) AddTask(ctx context.Context, req *crontaskv1.AddTaskRequest) (*crontaskv1.AddTaskResponse, error) {
	err := c.svc.AddTask(ctx, convertToDomain(req.Task))
	return &crontaskv1.AddTaskResponse{}, err
}

func convertToV(domainCronJob domain.Task) *crontaskv1.CronTask {
	return &crontaskv1.CronTask{
		Id:         domainCronJob.Id,
		Name:       domainCronJob.Name,
		Executor:   domainCronJob.Executor,
		Cfg:        domainCronJob.Cfg,
		Expression: domainCronJob.Expression,
		NextTime:   timestamppb.New(domainCronJob.NextTime),
	}
}

func convertToDomain(vCronJob *crontaskv1.CronTask) domain.Task {
	return domain.Task{
		Id:         vCronJob.GetId(),
		Name:       vCronJob.GetName(),
		Executor:   vCronJob.GetExecutor(),
		Cfg:        vCronJob.GetCfg(),
		Expression: vCronJob.GetExpression(),
		NextTime:   vCronJob.GetNextTime().AsTime(),
	}
}
