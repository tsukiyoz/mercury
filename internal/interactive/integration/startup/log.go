package startup

import "github.com/tsukiyo/mercury/pkg/logger"

func InitLog() logger.Logger {
	return &logger.NopLogger{}
}
