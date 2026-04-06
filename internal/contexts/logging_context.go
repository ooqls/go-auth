package contexts

import (
	"context"

	"go.uber.org/zap"
)

type LContext interface {
	context.Context
	L() *zap.Logger
}

type LoggingContext struct {
	context.Context
	l *zap.Logger
}

func NewLoggingContext(ctx context.Context, l *zap.Logger) *LoggingContext {
	return &LoggingContext{
		Context: ctx,
		l:       l,
	}
}

func (c *LoggingContext) L() *zap.Logger {
	return c.l
}
