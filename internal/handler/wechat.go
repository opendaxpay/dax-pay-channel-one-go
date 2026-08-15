package handler

import (
	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/service"
)

// 微信直连 / ISV 入口；业务错误统一经 writeErr 写出（见 alipay.go）。

// WechatPay：POST /channel/wechat/pay
func WechatPay(c *gin.Context) {
	var req dto.PayReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.DirectPay(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
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
		writeErr(c, err)
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
		writeErr(c, err)
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
		writeErr(c, err)
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
		writeErr(c, err)
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

// WechatTransfer：POST /channel/wechat/transfer
func WechatTransfer(c *gin.Context) {
	var req dto.TransferReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Transfer(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatTransferSync：POST /channel/wechat/transfer-sync
func WechatTransferSync(c *gin.Context) {
	var req dto.TransferReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.TransferSync(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatAlloc：POST /channel/wechat/alloc
func WechatAlloc(c *gin.Context) {
	var req dto.AllocReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.Alloc(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatAllocSync：POST /channel/wechat/alloc-sync
func WechatAllocSync(c *gin.Context) {
	var req dto.AllocReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.AllocSync(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}

// WechatCallbackParseTransfer：POST /channel/wechat/callback/parse-transfer
func WechatCallbackParseTransfer(c *gin.Context) {
	var req dto.CallbackParseReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	writeOK(c, service.ParseTransferCallback(c.Request.Context(), &req))
}

// WechatIsvPay：POST /channel/wechat/isv/pay
func WechatIsvPay(c *gin.Context) {
	var req dto.PayReq
	if !middleware.BindJSON(c, &req) {
		return
	}
	data, err := service.IsvPay(c.Request.Context(), &req)
	if err != nil {
		writeErr(c, err)
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
		writeErr(c, err)
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
		writeErr(c, err)
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
		writeErr(c, err)
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
		writeErr(c, err)
		return
	}
	writeOK(c, data)
}
