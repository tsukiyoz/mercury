package startup

import "github.com/tsukaychan/webook/pkg/logger"

func InitLog() logger.Logger {
	return &logger.NopLogger{}
}
