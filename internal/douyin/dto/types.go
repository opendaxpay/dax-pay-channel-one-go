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
	Allocation  *bool                 `json:"allocation"`
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

// ReportInfo：转账场景报备信息(transfer_scene_report_infos)
type ReportInfo struct {
	InfoType    string `json:"infoType"`
	InfoContent string `json:"infoContent"`
}

// TransferReq：转账发起/同步共用请求(对标 DouyinTransferReq)
//
// 发起时 amount/openid(或 phoneNumber)/scene 必填; 同步按 outBillNo(通道单号)或 transferNo(平台单号)反查。
type TransferReq struct {
	OutBillNo    string                `json:"outBillNo"`
	TransferNo   string                `json:"transferNo"`
	Amount       jsonx.Int64String     `json:"amount"`
	Openid       string                `json:"openid"`
	PhoneNumber  string                `json:"phoneNumber"`
	Scene        string                `json:"scene"`
	UserName     string                `json:"userName"`
	Remark       string                `json:"remark"`
	Perception   string                `json:"perception"`
	ReportInfos  []ReportInfo          `json:"reportInfos"`
	NotifyURL    string                `json:"notifyUrl"`
	Credential   *douyin.SdkCredential `json:"credential"`
}

// TransferResp：转账响应(对标 DouyinTransferResp)
type TransferResp struct {
	TransferBillNo string `json:"transferBillNo,omitempty"`
	State          string `json:"state,omitempty"`
	FailReason     string `json:"failReason,omitempty"`
}

// ReceiverInfo：分账接收方
type ReceiverInfo struct {
	Type    string            `json:"type"`
	Account string            `json:"account"`
	Name    string            `json:"name"`
	Amount  jsonx.Int64String `json:"amount"`
}

// AllocReq：分账发起/同步共用请求(对标 DouyinAllocReq)
type AllocReq struct {
	OutTradeNo        string                `json:"outTradeNo"`
	TradeNo           string                `json:"tradeNo"`
	ReceiverInfoDtos  []ReceiverInfo        `json:"receiverInfoDtos"`
	NotifyURL         string                `json:"notifyUrl"`
	Credential        *douyin.SdkCredential `json:"credential"`
}

// ReceiverSplitResult：分账逐明细结果
type ReceiverSplitResult struct {
	Account    string         `json:"account,omitempty"`
	Amount     *jsonx.Int64String `json:"amount,omitempty"`
	SplitStatus string        `json:"splitStatus,omitempty"`
	FailReason string         `json:"failReason,omitempty"`
	FinishTime string         `json:"finishTime,omitempty"`
}

// AllocResp：分账响应(对标 DouyinAllocResp)
type AllocResp struct {
	OrderId                 string                 `json:"orderId,omitempty"`
	Status                  string                 `json:"status,omitempty"`
	ReceiverSplitResultDtos []ReceiverSplitResult  `json:"receiverSplitResultDtos,omitempty"`
	ErrorCode               string                 `json:"errorCode,omitempty"`
	ErrorMsg                string                 `json:"errorMsg,omitempty"`
}

// TransferCallbackParseResp：转账回调解析响应(对标 DouyinTransferCallbackParseResp)
//
// 抖音转账通知仅含 order_id(通道转账单号), 不含商户单号。
type TransferCallbackParseResp struct {
	Verified           bool   `json:"verified"`
	TransferBillNo     string `json:"transferBillNo,omitempty"`
	TransferState      string `json:"transferState,omitempty"`
	TransferStatusDesc string `json:"transferStatusDesc,omitempty"`
	SuccessTime        string `json:"successTime,omitempty"`
}

// AllocCallbackReceiverResult：分账回调逐明细结果
type AllocCallbackReceiverResult struct {
	Account    string `json:"account,omitempty"`
	SplitStatus string `json:"splitStatus,omitempty"`
	FailReason string `json:"failReason,omitempty"`
	FinishTime string `json:"finishTime,omitempty"`
}

// AllocCallbackParseResp：分账回调解析响应(对标 DouyinAllocCallbackParseResp)
type AllocCallbackParseResp struct {
	Verified          bool                          `json:"verified"`
	OrderId           string                        `json:"orderId,omitempty"`
	State             string                        `json:"state,omitempty"`
	SplitFinishTime   string                        `json:"splitFinishTime,omitempty"`
	ReceiverResults   []AllocCallbackReceiverResult `json:"receiverResults,omitempty"`
}
