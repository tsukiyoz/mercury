package app

import (
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
