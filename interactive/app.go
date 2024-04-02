package main

import (
	"github.com/tsukaychan/webook/pkg/grpcx"
	"github.com/tsukaychan/webook/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
	// cron      *cron.Cron
}
