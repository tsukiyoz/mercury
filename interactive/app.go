package main

import (
	"github.com/lazywoo/mercury/pkg/ginx"
	"github.com/lazywoo/mercury/pkg/grpcx"
	"github.com/lazywoo/mercury/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	web       *ginx.Server
	consumers []saramax.Consumer
	// cron      *cron.Cron
}
