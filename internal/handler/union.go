package handler

import (
	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/union/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/union/service"
)

// 云闪付入口；业务错误统一经 writeErr 写出（见 alipay.go）。

// UnionPay：POST /channel/union/pay
func UnionPay(c *gin.Context) {
	var req dto.PayReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Pay(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}

// UnionSync：POST /channel/union/sync
func UnionSync(c *gin.Context) {
	var req dto.SyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Sync(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}

// UnionClose：POST /channel/union/close
func UnionClose(c *gin.Context) {
	var req dto.CloseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Close(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}

// UnionRefund：POST /channel/union/refund
func UnionRefund(c *gin.Context) {
	var req dto.RefundReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Refund(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}

// UnionRefundSync：POST /channel/union/refund-sync
func UnionRefundSync(c *gin.Context) {
	var req dto.RefundSyncReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.RefundSync(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}

// UnionCallbackParsePay：POST /channel/union/callback/parse-pay
func UnionCallbackParsePay(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParsePayCallback(c.Request.Context(), &req))
}

// UnionCallbackParseRefund：POST /channel/union/callback/parse-refund
func UnionCallbackParseRefund(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParseRefundCallback(c.Request.Context(), &req))
}
