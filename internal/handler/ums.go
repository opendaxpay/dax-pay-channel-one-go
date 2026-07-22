package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/result"
	"daxpay.open/dax-pay-channel-one-go/internal/ums"
	"daxpay.open/dax-pay-channel-one-go/internal/ums/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/ums/service"
)

func writeUmsErr(c *gin.Context, err error) {
	ctx := c.Request.Context()
	if be, ok := err.(*ums.BizError); ok {
		c.JSON(http.StatusOK, result.Fail(be.Code, ums.LocalizedMsg(ctx, be)))
		return
	}
	c.JSON(http.StatusOK, result.Fail(
		errcode.SDKCallFailed.Code,
		errcode.SDKCallFailed.MessageWithDetail(ctx, err.Error()),
	))
}

// UmsPay：POST /channel/ums/pay
func UmsPay(c *gin.Context) {
	var req dto.PayReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Pay(c.Request.Context(), &req)
	if err != nil {
		writeUmsErr(c, err)
		return
	}
	writeOK(c, data)
}

// UmsSync：POST /channel/ums/sync
func UmsSync(c *gin.Context) {
	var req dto.SyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Sync(c.Request.Context(), &req)
	if err != nil {
		writeUmsErr(c, err)
		return
	}
	writeOK(c, data)
}

// UmsClose：POST /channel/ums/close
func UmsClose(c *gin.Context) {
	var req dto.CloseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Close(c.Request.Context(), &req)
	if err != nil {
		writeUmsErr(c, err)
		return
	}
	writeOK(c, data)
}

// UmsRefund：POST /channel/ums/refund
func UmsRefund(c *gin.Context) {
	var req dto.RefundReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Refund(c.Request.Context(), &req)
	if err != nil {
		writeUmsErr(c, err)
		return
	}
	writeOK(c, data)
}

// UmsRefundSync：POST /channel/ums/refund-sync
func UmsRefundSync(c *gin.Context) {
	var req dto.RefundSyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.RefundSync(c.Request.Context(), &req)
	if err != nil {
		writeUmsErr(c, err)
		return
	}
	writeOK(c, data)
}

// UmsCallbackParsePay：POST /channel/ums/callback/parse-pay
func UmsCallbackParsePay(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParsePayCallback(c.Request.Context(), &req))
}

// UmsCallbackParseRefund：POST /channel/ums/callback/parse-refund
func UmsCallbackParseRefund(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParseRefundCallback(c.Request.Context(), &req))
}
