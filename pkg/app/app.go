package app

import (
	"github.com/robfig/cron/v3"
	"golang.org/x/sync/errgroup"

	"github.com/lazywoo/mercury/pkg/ginx"
	"github.com/lazywoo/mercury/pkg/grpcx"
	"github.com/lazywoo/mercury/pkg/saramax"
)

type App struct {
	GRPCServer *grpcx.Server
	WebServer  *ginx.Server
	Consumers  []saramax.Consumer
	Cron       *cron.Cron
}

func (a *App) Run() error {
	var eg errgroup.Group
	if a.GRPCServer != nil {
		eg.Go(func() error {
			return a.GRPCServer.Serve()
		})
	}

	if a.WebServer != nil {
		eg.Go(func() error {
			return a.WebServer.Start()
		})
	}

	if len(a.Consumers) > 0 {
		for _, consumer := range a.Consumers {
			eg.Go(func() error {
				return consumer.Start()
			})
		}
	}

	if a.Cron != nil {
		eg.Go(func() error {
			a.Cron.Start()
			return nil
		})
	}

	return eg.Wait()
}
