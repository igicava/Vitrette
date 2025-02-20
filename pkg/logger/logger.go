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
