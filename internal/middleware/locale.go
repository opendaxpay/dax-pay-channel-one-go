package middleware

import (
	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/i18n"
)

const HeaderAcceptLanguage = "Accept-Language"
const HeaderXClientCode = "x-client-code"

const ctxClientCode = "xClientCode"

// Locale：解析 Accept-Language → 资源 locale，写入 request context（国际化入口）
func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := i18n.ParseAcceptLanguage(c.GetHeader(HeaderAcceptLanguage))
		ctx := i18n.WithLocale(c.Request.Context(), locale)
		c.Request = c.Request.WithContext(ctx)
		if code := c.GetHeader(HeaderXClientCode); code != "" {
			c.Set(ctxClientCode, code)
		}
		c.Next()
	}
}
