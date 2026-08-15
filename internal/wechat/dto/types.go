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
	Allocation    *bool                 `json:"allocation"`
	Credential    *wechat.SdkCredential `json:"credential"`
}

// PayResp：下单响应（对标 WechatPayResp）
type PayResp struct {
	OutTradeNo    string                `json:"outTradeNo,omitempty"`
	TransactionID string                `json:"transactionId,omitempty"`
	PayBody       string                `json:"payBody,omitempty"`
	PayBodyType   PayBodyType           `json:"payBodyType,omitempty"`
	Complete      bool                  `json:"complete"`
	FinishTime    *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
	TotalAmount   *jsonx.Int64String    `json:"totalAmount,omitempty"`
	PayerTotal    *jsonx.Int64String    `json:"payerTotal,omitempty"`
	OpenID        string                `json:"openId,omitempty"`
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

// ReportInfo：转账场景报备信息(transfer_scene_report_infos)
type ReportInfo struct {
	InfoType    string `json:"infoType"`
	InfoContent string `json:"infoContent"`
}

// TransferReq：转账发起/同步共用请求(对标 WechatTransferReq)
//
// 发起时 amount/openid/scene 必填; 同步按 outBillNo(通道单号)或 transferNo(平台单号)反查。
type TransferReq struct {
	OutBillNo  string                `json:"outBillNo"`
	TransferNo string                `json:"transferNo"`
	Amount     jsonx.Int64String     `json:"amount"`
	Openid     string                `json:"openid"`
	Scene      string                `json:"scene"`
	UserName   string                `json:"userName"`
	Remark     string                `json:"remark"`
	NotifyURL  string                `json:"notifyUrl"`
	ReportInfos []ReportInfo         `json:"reportInfos"`
	Credential *wechat.SdkCredential `json:"credential"`
}

// TransferResp：转账响应(对标 WechatTransferResp)
type TransferResp struct {
	Complete       bool                  `json:"complete"`
	TransferBillNo string                `json:"transferBillNo,omitempty"`
	PackageInfo    string                `json:"packageInfo,omitempty"`
	State          string                `json:"state,omitempty"`
	FinishTime     *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
	FailReason     string                `json:"failReason,omitempty"`
}

// Receiver：分账接收方
type Receiver struct {
	Type        string            `json:"type"`
	Account     string            `json:"account"`
	Name        string            `json:"name"`
	Amount      jsonx.Int64String `json:"amount"`
	Description string            `json:"description"`
}

// AllocReq：分账发起/同步共用请求(对标 WechatAllocReq)
type AllocReq struct {
	OutOrderNo    string                `json:"outOrderNo"`
	TransactionID string                `json:"transactionId"`
	Receivers     []Receiver            `json:"receivers"`
	Credential    *wechat.SdkCredential `json:"credential"`
}

// ReceiverResult：分账逐明细结果
type ReceiverResult struct {
	Account    string                `json:"account,omitempty"`
	Amount     *jsonx.Int64String    `json:"amount,omitempty"`
	Result     string                `json:"result,omitempty"`
	FailReason string                `json:"failReason,omitempty"`
	FinishTime *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
}

// AllocResp：分账响应(对标 WechatAllocResp)
type AllocResp struct {
	TransactionID string           `json:"transactionId,omitempty"`
	State         string           `json:"state,omitempty"`
	Receivers     []ReceiverResult `json:"receivers,omitempty"`
	ErrorCode     string           `json:"errorCode,omitempty"`
	ErrorMsg      string           `json:"errorMsg,omitempty"`
}

// TransferCallbackParseResp：转账回调解析响应(对标 WechatTransferCallbackParseResp)
type TransferCallbackParseResp struct {
	Verified       bool   `json:"verified"`
	OutBillNo      string `json:"outBillNo,omitempty"`
	TransferBillNo string `json:"transferBillNo,omitempty"`
	TransferState  string `json:"transferState,omitempty"`
	FailReason     string `json:"failReason,omitempty"`
	UpdateTime     string `json:"updateTime,omitempty"`
}
