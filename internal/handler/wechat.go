package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/result"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/service"
)

func writeWechatErr(c *gin.Context, err error) {
	ctx := c.Request.Context()
	if be, ok := err.(*wechat.BizError); ok {
		c.JSON(http.StatusOK, result.Fail(be.Code, wechat.LocalizedMsg(ctx, be)))
		return
	}
	c.JSON(http.StatusOK, result.Fail(
		errcode.SDKCallFailed.Code,
		errcode.SDKCallFailed.MessageWithDetail(ctx, err.Error()),
	))
}

// WechatPay：POST /channel/wechat/pay
func WechatPay(c *gin.Context) {
	var req dto.PayReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.DirectPay(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatSync：POST /channel/wechat/sync
func WechatSync(c *gin.Context) {
	var req dto.SyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.DirectSync(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatClose：POST /channel/wechat/close
func WechatClose(c *gin.Context) {
	var req dto.CloseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.DirectClose(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatRefund：POST /channel/wechat/refund
func WechatRefund(c *gin.Context) {
	var req dto.RefundReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.DirectRefund(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatRefundSync：POST /channel/wechat/refund-sync
func WechatRefundSync(c *gin.Context) {
	var req dto.RefundSyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.DirectRefundSync(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatCallbackParsePay：POST /channel/wechat/callback/parse-pay
func WechatCallbackParsePay(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParsePayCallback(c.Request.Context(), &req))
}

// WechatCallbackParseRefund：POST /channel/wechat/callback/parse-refund
func WechatCallbackParseRefund(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParseRefundCallback(c.Request.Context(), &req))
}

// WechatIsvPay：POST /channel/wechat/isv/pay
func WechatIsvPay(c *gin.Context) {
	var req dto.PayReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.IsvPay(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatIsvSync：POST /channel/wechat/isv/sync
func WechatIsvSync(c *gin.Context) {
	var req dto.SyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.IsvSync(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatIsvClose：POST /channel/wechat/isv/close
func WechatIsvClose(c *gin.Context) {
	var req dto.CloseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.IsvClose(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatIsvRefund：POST /channel/wechat/isv/refund
func WechatIsvRefund(c *gin.Context) {
	var req dto.RefundReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.IsvRefund(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatIsvRefundSync：POST /channel/wechat/isv/refund-sync
func WechatIsvRefundSync(c *gin.Context) {
	var req dto.RefundSyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.IsvRefundSync(c.Request.Context(), &req)
	if err != nil {
		writeWechatErr(c, err)
		return
	}
	writeOK(c, data)
}
