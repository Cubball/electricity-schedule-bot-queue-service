package logger

import (
	"context"
	"log/slog"
)

const TraceIdContextKey = "traceId"

type TraceIdHandler struct {
	wrappedHandler slog.Handler
}

func NewTraceIdHandler(wrappedHandler slog.Handler) slog.Handler {
	return TraceIdHandler{
		wrappedHandler: wrappedHandler,
	}
}

func (h TraceIdHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.wrappedHandler.Enabled(ctx, level)
}

func (h TraceIdHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewTraceIdHandler(h.wrappedHandler.WithAttrs(attrs))
}

func (h TraceIdHandler) WithGroup(name string) slog.Handler {
	return NewTraceIdHandler(h.wrappedHandler.WithGroup(name))
}

func (h TraceIdHandler) Handle(ctx context.Context, record slog.Record) error {
	traceId, ok := ctx.Value(TraceIdContextKey).(string)
	if ok {
		record.Add(TraceIdContextKey, traceId)
	}

	return h.wrappedHandler.Handle(ctx, record)
}
