package middleware

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer：进程内 TracerProvider + W3C TraceContext；默认不启用 OTLP 导出
func InitTracer(ctx context.Context, serviceName string, sampleRatio float64) (func(context.Context) error, error) {
	if sampleRatio <= 0 {
		sampleRatio = 1.0
	}
	if sampleRatio > 1 {
		sampleRatio = 1
	}

	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRatio))),
		sdktrace.WithResource(res),
		// 无 SpanExporter：仅内存中创建 span，供日志与 x-trace-id 关联（对标 otlp.enabled=false）
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	slog.Info("otel tracer initialized", "service", serviceName, "sampleRatio", sampleRatio, "otlp", false)

	return tp.Shutdown, nil
}
