package middleware

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// LoggerWithTrace：返回带 traceId/spanId 的 slog.Logger（日志链路关联）
func LoggerWithTrace(ctx context.Context) *slog.Logger {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if !sc.IsValid() {
		return slog.Default()
	}
	return slog.Default().With(
		"traceId", sc.TraceID().String(),
		"spanId", sc.SpanID().String(),
	)
}
