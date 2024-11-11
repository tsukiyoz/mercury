package app

import (
	"sync"

	"github.com/robfig/cron/v3"

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

func (a *App) Run() <-chan struct{} {
	close := make(chan struct{})
	var wg sync.WaitGroup
	if a.GRPCServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.GRPCServer.Serve(); err != nil {
				panic(err)
			}
		}()
	}

	if a.WebServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.WebServer.Start(); err != nil {
				panic(err)
			}
		}()
	}

	if len(a.Consumers) > 0 {
		wg.Add(len(a.Consumers))
		for _, consumer := range a.Consumers {
			go func(c saramax.Consumer) {
				defer wg.Done()
				if err := c.Start(); err != nil {
					panic(err)
				}
			}(consumer)
		}
	}

	if a.Cron != nil {
		a.Cron.Start()
	}

	go func() {
		wg.Wait()
		close <- struct{}{}
	}()

	return close
}
