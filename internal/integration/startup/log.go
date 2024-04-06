package startup

import "github.com/tsukaychan/mercury/pkg/logger"

func InitLog() logger.Logger {
	return &logger.NopLogger{}
}
