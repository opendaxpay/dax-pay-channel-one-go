package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/result"
)

// AlipayStub：支付宝路径占位（首期无真实 SDK）
//
// 返回 HTTP 200 + code=10003 + 本地化 sdkCallFailedWithDetail，说明 Go 骨架未实现。
func AlipayStub(c *gin.Context) {
	ctx := c.Request.Context()
	middleware.LoggerWithTrace(ctx).Info("alipay stub hit", "path", c.FullPath())
	c.JSON(http.StatusOK, result.Fail(
		errcode.SDKCallFailed.Code,
		errcode.SDKCallFailed.MessageWithDetail(ctx, "Go skeleton not implemented"),
	))
}
