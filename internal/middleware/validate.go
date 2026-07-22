package middleware

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
	"daxpay.open/dax-pay-channel-one-go/internal/result"
)

// BindJSON：读取 body 并用 jsonx 解码（忽略未知字段）；失败则写 HTTP 200 + code=10007
//
// 返回 false 表示已写入错误响应，调用方应直接 return。
func BindJSON(c *gin.Context, dst any) bool {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		ctx := c.Request.Context()
		c.AbortWithStatusJSON(http.StatusOK, result.Fail(
			errcode.ValidateParams.Code,
			errcode.ValidateParams.Message(ctx),
		))
		return false
	}
	if err := jsonx.Unmarshal(body, dst); err != nil {
		ctx := c.Request.Context()
		c.AbortWithStatusJSON(http.StatusOK, result.Fail(
			errcode.ValidateParams.Code,
			errcode.ValidateParams.Message(ctx)+": "+err.Error(),
		))
		return false
	}
	return true
}
