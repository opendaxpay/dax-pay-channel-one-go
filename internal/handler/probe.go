package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/i18n"
	"daxpay.open/dax-pay-channel-one-go/internal/result"
)

// ProbeData：契约探测回显
type ProbeData struct {
	// 解析后的资源 locale
	ResolvedLocale string `json:"resolvedLocale"`
	// 当前 span 的 TraceID（与 x-trace-id 一致）
	TraceID string `json:"traceId"`
	SpanID  string `json:"spanId"`
}

// Probe：验证 i18n + tracing + DaxResult（国际化：success 文案随 Accept-Language 变化）
func Probe(c *gin.Context) {
	ctx := c.Request.Context()
	locale := i18n.LocaleFrom(ctx)
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	data := ProbeData{
		ResolvedLocale: locale,
		TraceID:        sc.TraceID().String(),
		SpanID:         sc.SpanID().String(),
	}
	// 成功文案：channel.error.success
	c.JSON(http.StatusOK, result.Ok(errcode.Success.Message(ctx), data))
}
