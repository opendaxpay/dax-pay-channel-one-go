package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// NormalizeTraceparent：归一化入站 traceparent 的 flags，兼容 OTel Java 注入的非标准 flags。
//
// 背景：主应用(micrometer-tracing-bridge-otel) 注入的 traceparent flags 为 03(带 W3C 保留位 bit1)，
// 而 OTel Go 1.37 严格校验 version 0 时 flags ≤ 2(trace_context.go: opts[0] > 2 即拒绝)，
// 导致 otelgin 提取失败、自创新 traceId，主应用 ↔ Go 通道链路断开。
//
// 本中间件在 otelgin 之前清除 flags 的非 sampled 位(仅保留 bit0)，让 otelgin 正常提取父 traceId。
// sampled 语义不变(主应用 flags=03 的 bit0=1 → 归一化为 01，仍为 sampled)。
//
// 必须注册在 [Otel] 之前。
func NormalizeTraceparent() gin.HandlerFunc {
	return func(c *gin.Context) {
		tp := c.GetHeader("traceparent")
		if tp != "" {
			// traceparent 格式: version-traceid-spanid-flags
			parts := strings.Split(tp, "-")
			if len(parts) == 4 && len(parts[3]) == 2 {
				if flags, err := strconv.ParseUint(parts[3], 16, 8); err == nil {
					// 仅保留 sampled 位(bit0)，清除 OTel Java 设入的 W3C 保留位
					normalized := flags & 1
					if normalized != flags {
						parts[3] = fmt.Sprintf("%02x", normalized)
						// 重写请求头，otelgin Extract 时拿到合规 flags
						c.Request.Header.Set("traceparent", strings.Join(parts, "-"))
					}
				}
			}
		}
		c.Next()
	}
}
