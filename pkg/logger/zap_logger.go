package logger

import "go.uber.org/zap"

var _ Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(logger *zap.Logger) Logger {
	return &ZapLogger{logger: logger}
}

func (l *ZapLogger) Info(msg string, args ...Field) {
	l.logger.Info(msg, l.toZapFields(args...)...)
}

func (l *ZapLogger) Debug(msg string, args ...Field) {
	l.logger.Debug(msg, l.toZapFields(args...)...)
}

func (l *ZapLogger) Warn(msg string, args ...Field) {
	l.logger.Debug(msg, l.toZapFields(args...)...)
}

func (l *ZapLogger) Error(msg string, args ...Field) {
	l.logger.Error(msg, l.toZapFields(args...)...)
}

func (l *ZapLogger) With(args ...Field) Logger {
	return &ZapLogger{
		logger: l.logger.With(l.toZapFields(args...)...),
	}
}

func (l *ZapLogger) toZapFields(args ...Field) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Value))
	}
	return res
}
