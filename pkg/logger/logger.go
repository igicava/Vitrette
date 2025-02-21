package logger

import (
	"context"
	"go.uber.org/zap"
)

func NewContext() context.Context {
	logger, _ := zap.NewProduction()
	ctx := context.WithValue(context.Background(), "logger", logger)
	return ctx
}

func GetLogger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value("logger").(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}
	return logger
}
