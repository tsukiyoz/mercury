package logger

import "go.uber.org/zap"

var _ Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(logger *zap.Logger) Logger {
	return &ZapLogger{logger: logger}
}

func (z *ZapLogger) Info(msg string, args ...Field) {
	z.logger.Info(msg, z.toZapFields(args...)...)
}

func (z *ZapLogger) Debug(msg string, args ...Field) {
	z.logger.Debug(msg, z.toZapFields(args...)...)
}

func (z *ZapLogger) Warn(msg string, args ...Field) {
	z.logger.Debug(msg, z.toZapFields(args...)...)
}

func (z *ZapLogger) Error(msg string, args ...Field) {
	z.logger.Error(msg, z.toZapFields(args...)...)
}

func (z *ZapLogger) toZapFields(args ...Field) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Value))
	}
	return res
}
