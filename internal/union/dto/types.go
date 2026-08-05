package dto

import (
	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
	"daxpay.open/dax-pay-channel-one-go/internal/union"
)

// PayReq：下单请求
type PayReq struct {
	OutTradeNo  string               `json:"outTradeNo"`
	Amount      jsonx.Int64String    `json:"amount"`
	Description string               `json:"description"`
	Method      PayMethod            `json:"method"`
	NotifyURL   string               `json:"notifyUrl"`
	AuthCode    string               `json:"authCode"`
	ClientIP    string               `json:"clientIp"`
	Credential  *union.SdkCredential `json:"credential"`
}

// PayResp：下单响应
type PayResp struct {
	OutTradeNo  string      `json:"outTradeNo,omitempty"`
	PayBody     string      `json:"payBody,omitempty"`
	PayBodyType PayBodyType `json:"payBodyType,omitempty"`
}

// SyncReq：查单
type SyncReq struct {
	OutTradeNo string               `json:"outTradeNo"`
	Method     PayMethod            `json:"method"`
	Credential *union.SdkCredential `json:"credential"`
}

// SyncResp：查单响应
type SyncResp struct {
	OutTradeNo  string `json:"outTradeNo,omitempty"`
	TradeStatus string `json:"tradeStatus,omitempty"`
	TotalAmount string `json:"totalAmount,omitempty"`
	RealAmount  string `json:"realAmount,omitempty"`
	PayTime     string `json:"payTime,omitempty"`
	QueryID     string `json:"queryId,omitempty"`
	BuyerID     string `json:"buyerId,omitempty"`
	ErrorMsg    string `json:"errorMsg,omitempty"`
}

// CloseReq：关单
type CloseReq struct {
	OutTradeNo string               `json:"outTradeNo"`
	QueryID    string               `json:"queryId"`
	Method     PayMethod            `json:"method"`
	Credential *union.SdkCredential `json:"credential"`
}

// CloseResp：关单响应
type CloseResp struct {
	OutTradeNo string `json:"outTradeNo,omitempty"`
}

// RefundReq：退款
type RefundReq struct {
	OutTradeNo   string               `json:"outTradeNo"`
	OrigQueryID  string               `json:"origQueryId"`
	OutRefundNo  string               `json:"outRefundNo"`
	RefundAmount jsonx.Int64String    `json:"refundAmount"`
	NotifyURL    string               `json:"notifyUrl"`
	Method       PayMethod            `json:"method"`
	Credential   *union.SdkCredential `json:"credential"`
}

// RefundResp：退款响应
type RefundResp struct {
	OutRefundNo  string `json:"outRefundNo,omitempty"`
	RefundStatus string `json:"refundStatus,omitempty"`
	FinishTime   string `json:"finishTime,omitempty"`
}

// RefundSyncReq：退款查询
type RefundSyncReq struct {
	OutRefundNo string               `json:"outRefundNo"`
	Method      PayMethod            `json:"method"`
	Credential  *union.SdkCredential `json:"credential"`
}

// RefundSyncResp：退款查询响应
type RefundSyncResp struct {
	OutRefundNo  string `json:"outRefundNo,omitempty"`
	RefundStatus string `json:"refundStatus,omitempty"`
	RefundAmount string `json:"refundAmount,omitempty"`
	FinishTime   string `json:"finishTime,omitempty"`
	ErrorMsg     string `json:"errorMsg,omitempty"`
}

// CallbackParseReq：回调验签解析
type CallbackParseReq struct {
	Credential *union.SdkCredential `json:"credential"`
	Params     map[string]string    `json:"params"`
}

// CallbackParseResp：回调解析结果
type CallbackParseResp struct {
	Verified    bool   `json:"verified"`
	TradeType   string `json:"tradeType,omitempty"`
	OutTradeNo  string `json:"outTradeNo,omitempty"`
	OutRefundNo string `json:"outRefundNo,omitempty"`
	TradeStatus string `json:"tradeStatus,omitempty"`
	Amount      string `json:"amount,omitempty"`
	RealAmount  string `json:"realAmount,omitempty"`
	FinishTime  string `json:"finishTime,omitempty"`
	QueryID     string `json:"queryId,omitempty"`
	BuyerID     string `json:"buyerId,omitempty"`
}
