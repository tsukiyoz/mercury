package main

import (
	"github.com/tsukaychan/mercury/pkg/ginx"
	"github.com/tsukaychan/mercury/pkg/grpcx"
	"github.com/tsukaychan/mercury/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	web       *ginx.Server
	consumers []saramax.Consumer
	// cron      *cron.Cron
}
