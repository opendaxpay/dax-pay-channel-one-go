package dto

import (
	"daxpay.open/dax-pay-channel-one-go/internal/douyin"
	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
)

// PayReq：下单（对标 Boot DouyinPayReq）
type PayReq struct {
	OutTradeNo  string                `json:"outTradeNo"`
	Amount      jsonx.Int64String     `json:"amount"`
	Description string                `json:"description"`
	Method      PayMethod             `json:"method"`
	OpenID      string                `json:"openId"`
	ClientIP    string                `json:"clientIp"`
	ExpiredTime *jsonx.OffsetDateTime `json:"expiredTime"`
	NotifyURL   string                `json:"notifyUrl"`
	Credential  *douyin.SdkCredential `json:"credential"`
}

// PayResp：下单响应
type PayResp struct {
	OutTradeNo  string      `json:"outTradeNo,omitempty"`
	PayBody     string      `json:"payBody,omitempty"`
	PayBodyType PayBodyType `json:"payBodyType,omitempty"`
}

// CloseReq：关单
type CloseReq struct {
	OutTradeNo string                `json:"outTradeNo"`
	Credential *douyin.SdkCredential `json:"credential"`
}

// CloseResp：关单响应
type CloseResp struct {
	OutTradeNo string `json:"outTradeNo,omitempty"`
}

// RefundReq：退款
type RefundReq struct {
	OutTradeNo   string                `json:"outTradeNo"`
	OutRefundNo  string                `json:"outRefundNo"`
	RefundAmount jsonx.Int64String     `json:"refundAmount"`
	TotalAmount  jsonx.Int64String     `json:"totalAmount"`
	Reason       string                `json:"reason"`
	NotifyURL    string                `json:"notifyUrl"`
	Credential   *douyin.SdkCredential `json:"credential"`
}

// RefundResp：退款响应（finishTime 为抖音 RFC3339 原串）
type RefundResp struct {
	OutRefundNo  string `json:"outRefundNo,omitempty"`
	RefundID     string `json:"refundId,omitempty"`
	RefundStatus string `json:"refundStatus,omitempty"`
	FinishTime   string `json:"finishTime,omitempty"`
}

// SyncReq：查单
type SyncReq struct {
	OutTradeNo string                `json:"outTradeNo"`
	Credential *douyin.SdkCredential `json:"credential"`
}

// SyncResp：查单响应
type SyncResp struct {
	OutTradeNo    string             `json:"outTradeNo,omitempty"`
	TransactionID string             `json:"transactionId,omitempty"`
	TradeState    string             `json:"tradeState,omitempty"`
	TotalAmount   *jsonx.Int64String `json:"totalAmount,omitempty"`
	OpenID        string             `json:"openid,omitempty"`
	SuccessTime   string             `json:"successTime,omitempty"`
	ErrorCode     string             `json:"errorCode,omitempty"`
	ErrorMsg      string             `json:"errorMsg,omitempty"`
}

// RefundSyncReq：退款查询
type RefundSyncReq struct {
	OutRefundNo string                `json:"outRefundNo"`
	Credential  *douyin.SdkCredential `json:"credential"`
}

// RefundSyncResp：退款查询响应
type RefundSyncResp struct {
	OutRefundNo  string             `json:"outRefundNo,omitempty"`
	RefundID     string             `json:"refundId,omitempty"`
	RefundStatus string             `json:"refundStatus,omitempty"`
	RefundAmount *jsonx.Int64String `json:"refundAmount,omitempty"`
	FinishTime   string             `json:"finishTime,omitempty"`
	ErrorCode    string             `json:"errorCode,omitempty"`
	ErrorMsg     string             `json:"errorMsg,omitempty"`
}

// CallbackParseReq：回调验签解析（主应用转发 header + body）
type CallbackParseReq struct {
	Credential *douyin.SdkCredential `json:"credential"`
	Body       string                `json:"body"`
	Serial     string                `json:"serial"`
	Nonce      string                `json:"nonce"`
	Signature  string                `json:"signature"`
	Timestamp  string                `json:"timestamp"`
}

// CallbackParseResp：回调解析结果
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
	OpenID        string             `json:"openid,omitempty"`
	Verified      bool               `json:"verified"`
}
