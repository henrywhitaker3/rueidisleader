package rueidisleader

import "go.uber.org/zap"

type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
}

type nilLogger struct{}

func (n nilLogger) Info(msg string, args ...any)  {}
func (n nilLogger) Debug(msg string, args ...any) {}
func (n nilLogger) Error(msg string, args ...any) {}

type zapSugaredWrapper struct {
	zap *zap.SugaredLogger
}

func ZapLogger(z *zap.SugaredLogger) Logger {
	return &zapSugaredWrapper{
		zap: z,
	}
}

func (z *zapSugaredWrapper) Info(msg string, args ...any) {
	z.zap.Infow(msg, args...)
}

func (z *zapSugaredWrapper) Error(msg string, args ...any) {
	z.zap.Errorw(msg, args...)
}

func (z *zapSugaredWrapper) Debug(msg string, args ...any) {
	z.zap.Debugw(msg, args...)
}
