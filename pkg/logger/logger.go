package logger

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

const (
	Key = "logger"

	RequestID = "request_id"
)

type Logger struct {
	l *zap.Logger
}

func NewContext(ctx context.Context) (context.Context, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, Key, &Logger{logger})
	return ctx, nil
}

func GetLogger(ctx context.Context) *Logger {
	if _, ok := ctx.Value(Key).(*Logger); !ok {
		unknownCtx, _ := NewContext(ctx)
		return unknownCtx.Value(Key).(*Logger)
	}
	return ctx.Value(Key).(*Logger)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.l.Info(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.l.Fatal(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.l.Error(msg, fields...)
}

func Interceptor(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	next grpc.UnaryHandler,
) (any, error) {

	guid := uuid.New().String()
	ctx = context.WithValue(ctx, RequestID, guid)

	GetLogger(ctx).Info(ctx,
		"request", zap.String("method", info.FullMethod),
		zap.Time("request time", time.Now()),
	)

	return next(ctx, req)
}
