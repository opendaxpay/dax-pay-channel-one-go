package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"

	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/result"
)

// 响应头名称（与主应用 WebHeaderCode.X_TRACE_ID 一致）
const HeaderXTraceID = "x-trace-id"

// Otel：W3C traceparent 传播（otelgin）
func Otel(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

// TraceIDHeader：将当前 span 的 TraceID 写入 x-trace-id 响应头（对标 TraceIdFilter）
//
// 必须在 handler 写 body 之前设置，否则 Gin 已刷响应头会丢字段。
func TraceIDHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		sc := span.SpanContext()
		if sc.HasTraceID() {
			c.Header(HeaderXTraceID, sc.TraceID().String())
		}
		c.Next()
	}
}

// Recovery：panic → HTTP 200 + DaxResult code=10006
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				ctx := c.Request.Context()
				slog.ErrorContext(ctx, "panic recovered", "err", r)
				c.AbortWithStatusJSON(http.StatusOK, result.Fail(
					errcode.SystemError.Code,
					errcode.SystemError.Message(ctx),
				))
			}
		}()
		c.Next()
	}
}
