package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/service"
	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/result"
)

func writeBizErr(c *gin.Context, err error) {
	ctx := c.Request.Context()
	if be, ok := err.(*alipay.BizError); ok {
		c.JSON(http.StatusOK, result.Fail(be.Code, alipay.LocalizedMsg(ctx, be)))
		return
	}
	c.JSON(http.StatusOK, result.Fail(
		errcode.SDKCallFailed.Code,
		errcode.SDKCallFailed.MessageWithDetail(ctx, err.Error()),
	))
}

func writeOK(c *gin.Context, data any) {
	ctx := c.Request.Context()
	c.JSON(http.StatusOK, result.Ok(errcode.Success.Message(ctx), data))
}

// AlipayPay：POST /channel/alipay/pay
func AlipayPay(c *gin.Context) {
	var req dto.PayReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Pay(c.Request.Context(), &req)
	if err != nil {
		writeBizErr(c, err)
		return
	}
	writeOK(c, data)
}

// AlipaySync：POST /channel/alipay/sync
func AlipaySync(c *gin.Context) {
	var req dto.SyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Sync(c.Request.Context(), &req)
	if err != nil {
		writeBizErr(c, err)
		return
	}
	writeOK(c, data)
}

// AlipayClose：POST /channel/alipay/close
func AlipayClose(c *gin.Context) {
	var req dto.CloseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Close(c.Request.Context(), &req)
	if err != nil {
		writeBizErr(c, err)
		return
	}
	writeOK(c, data)
}

// AlipayRefund：POST /channel/alipay/refund
func AlipayRefund(c *gin.Context) {
	var req dto.RefundReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Refund(c.Request.Context(), &req)
	if err != nil {
		writeBizErr(c, err)
		return
	}
	writeOK(c, data)
}

// AlipayRefundSync：POST /channel/alipay/refund-sync
func AlipayRefundSync(c *gin.Context) {
	var req dto.RefundSyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.RefundSync(c.Request.Context(), &req)
	if err != nil {
		writeBizErr(c, err)
		return
	}
	writeOK(c, data)
}

// AlipayCallbackParsePay：POST /channel/alipay/callback/parse-pay
func AlipayCallbackParsePay(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParsePayCallback(c.Request.Context(), &req))
}

// AlipayCallbackParseRefund：POST /channel/alipay/callback/parse-refund
func AlipayCallbackParseRefund(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParseRefundCallback(c.Request.Context(), &req))
}

// AlipayAppAuthToken：POST /channel/alipay/auth/app-token
func AlipayAppAuthToken(c *gin.Context) {
	var req dto.AppAuthTokenReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.ExchangeAppAuthToken(c.Request.Context(), &req)
	if err != nil {
		writeBizErr(c, err)
		return
	}
	writeOK(c, data)
}
