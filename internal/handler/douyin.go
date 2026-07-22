package handler

import (
	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/douyin/service"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// 抖音入口；业务错误统一经 writeErr 写出（见 alipay.go）。

// DouyinPay：POST /channel/douyin/pay
func DouyinPay(c *gin.Context) {
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

// DouyinSync：POST /channel/douyin/sync
func DouyinSync(c *gin.Context) {
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

// DouyinClose：POST /channel/douyin/close
func DouyinClose(c *gin.Context) {
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

// DouyinRefund：POST /channel/douyin/refund
func DouyinRefund(c *gin.Context) {
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

// DouyinRefundSync：POST /channel/douyin/refund-sync
func DouyinRefundSync(c *gin.Context) {
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

// DouyinCallbackParsePay：POST /channel/douyin/callback/parse-pay
func DouyinCallbackParsePay(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParsePayCallback(c.Request.Context(), &req))
}

// DouyinCallbackParseRefund：POST /channel/douyin/callback/parse-refund
func DouyinCallbackParseRefund(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParseRefundCallback(c.Request.Context(), &req))
}
