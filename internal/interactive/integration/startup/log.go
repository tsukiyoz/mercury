package startup

import "github.com/lazywoo/mercury/pkg/logger"

func InitLog() logger.Logger {
	return &logger.NopLogger{}
}
