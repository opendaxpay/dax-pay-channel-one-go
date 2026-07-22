package dto

import (
	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat"
)

// PayReq：下单请求（对标 WechatPayReq）
type PayReq struct {
	OutTradeNo    string                `json:"outTradeNo"`
	Amount        jsonx.Int64String     `json:"amount"`
	Description   string                `json:"description"`
	Method        PayMethod             `json:"method"`
	ExpireTime    *jsonx.OffsetDateTime `json:"expireTime"`
	NotifyURL     string                `json:"notifyUrl"`
	Attach        string                `json:"attach"`
	OpenID        string                `json:"openId"`
	AuthCode      string                `json:"authCode"`
	PayerClientIp string                `json:"payerClientIp"`
	WapURL        string                `json:"wapUrl"`
	WapName       string                `json:"wapName"`
	Credential    *wechat.SdkCredential `json:"credential"`
}

// PayResp：下单响应（对标 WechatPayResp）
type PayResp struct {
	OutTradeNo    string                 `json:"outTradeNo,omitempty"`
	TransactionID string                 `json:"transactionId,omitempty"`
	PayBody       string                 `json:"payBody,omitempty"`
	PayBodyType   PayBodyType            `json:"payBodyType,omitempty"`
	Complete      bool                   `json:"complete"`
	FinishTime    *jsonx.OffsetDateTime  `json:"finishTime,omitempty"`
	TotalAmount   *jsonx.Int64String     `json:"totalAmount,omitempty"`
	PayerTotal    *jsonx.Int64String     `json:"payerTotal,omitempty"`
	OpenID        string                 `json:"openId,omitempty"`
}

// SyncReq：查单请求
type SyncReq struct {
	OutTradeNo    string                `json:"outTradeNo"`
	TransactionID string                `json:"transactionId"`
	Credential    *wechat.SdkCredential `json:"credential"`
}

// SyncResp：查单响应
type SyncResp struct {
	TradeState     string                `json:"tradeState,omitempty"`
	TradeStateDesc string                `json:"tradeStateDesc,omitempty"`
	TransactionID  string                `json:"transactionId,omitempty"`
	OutTradeNo     string                `json:"outTradeNo,omitempty"`
	SuccessTime    *jsonx.OffsetDateTime `json:"successTime,omitempty"`
	TotalAmount    *jsonx.Int64String    `json:"totalAmount,omitempty"`
	PayerTotal     *jsonx.Int64String    `json:"payerTotal,omitempty"`
	OpenID         string                `json:"openId,omitempty"`
}

// CloseReq：关单请求
type CloseReq struct {
	OutTradeNo    string                `json:"outTradeNo"`
	TransactionID string                `json:"transactionId"`
	Credential    *wechat.SdkCredential `json:"credential"`
}

// CloseResp：关单响应
type CloseResp struct {
	OutTradeNo    string `json:"outTradeNo,omitempty"`
	TransactionID string `json:"transactionId,omitempty"`
}

// RefundReq：退款请求
type RefundReq struct {
	OutTradeNo    string                `json:"outTradeNo"`
	TransactionID string                `json:"transactionId"`
	OutRefundNo   string                `json:"outRefundNo"`
	TotalAmount   jsonx.Int64String     `json:"totalAmount"`
	RefundAmount  jsonx.Int64String     `json:"refundAmount"`
	Reason        string                `json:"reason"`
	NotifyURL     string                `json:"notifyUrl"`
	Credential    *wechat.SdkCredential `json:"credential"`
}

// RefundResp：退款响应
type RefundResp struct {
	OutTradeNo    string                `json:"outTradeNo,omitempty"`
	TransactionID string                `json:"transactionId,omitempty"`
	OutRefundNo   string                `json:"outRefundNo,omitempty"`
	RefundID      string                `json:"refundId,omitempty"`
	Status        string                `json:"status,omitempty"`
	Complete      bool                  `json:"complete"`
	FinishTime    *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
	RefundAmount  *jsonx.Int64String    `json:"refundAmount,omitempty"`
	PayerRefund   *jsonx.Int64String    `json:"payerRefund,omitempty"`
}

// RefundSyncReq：退款查询请求
type RefundSyncReq struct {
	OutRefundNo string                `json:"outRefundNo"`
	Credential  *wechat.SdkCredential `json:"credential"`
}

// RefundSyncResp：退款查询响应
type RefundSyncResp struct {
	Status        string                `json:"status,omitempty"`
	RefundID      string                `json:"refundId,omitempty"`
	OutRefundNo   string                `json:"outRefundNo,omitempty"`
	TransactionID string                `json:"transactionId,omitempty"`
	OutTradeNo    string                `json:"outTradeNo,omitempty"`
	FinishTime    *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
	RefundAmount  *jsonx.Int64String    `json:"refundAmount,omitempty"`
	PayerRefund   *jsonx.Int64String    `json:"payerRefund,omitempty"`
}

// CallbackParseReq：回调验签解析请求（对标 WechatCallbackParseReq）
type CallbackParseReq struct {
	Credential *wechat.SdkCredential `json:"credential"`
	Body       string                `json:"body"`
	Serial     string                `json:"serial"`
	Nonce      string                `json:"nonce"`
	Signature  string                `json:"signature"`
	Timestamp  string                `json:"timestamp"`
}

// CallbackParseResp：回调验签解析响应（对标 WechatCallbackParseResp）
type CallbackParseResp struct {
	TradeType     string             `json:"tradeType,omitempty"`
	OutTradeNo    string             `json:"outTradeNo,omitempty"`
	TransactionID string             `json:"transactionId,omitempty"`
	OutRefundNo   string             `json:"outRefundNo,omitempty"`
	RefundID      string             `json:"refundId,omitempty"`
	TradeState    string             `json:"tradeState,omitempty"`
	RefundStatus  string             `json:"refundStatus,omitempty"`
	Amount        *jsonx.Int64String `json:"amount,omitempty"`
	SuccessTime   string             `json:"successTime,omitempty"`
	Openid        string             `json:"openid,omitempty"`
	Verified      bool               `json:"verified"`
}
