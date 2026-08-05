package dto

import (
	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
)

// PayReq：下单请求
type PayReq struct {
	OutTradeNo string                `json:"outTradeNo"`
	Amount     jsonx.Int64String     `json:"amount"`
	Subject    string                `json:"subject"`
	Body       string                `json:"body"`
	Method     PayMethod             `json:"method"`
	ExpireTime *jsonx.OffsetDateTime `json:"expireTime"`
	NotifyURL  string                `json:"notifyUrl"`
	AuthCode   string                `json:"authCode"`
	OpenID     string                `json:"openId"`
	Credential *alipay.SdkCredential `json:"credential"`
}

// PayResp：下单响应
type PayResp struct {
	OutTradeNo     string                `json:"outTradeNo,omitempty"`
	TradeNo        string                `json:"tradeNo,omitempty"`
	PayBody        string                `json:"payBody,omitempty"`
	PayBodyType    PayBodyType           `json:"payBodyType,omitempty"`
	Complete       bool                  `json:"complete"`
	FinishTime     *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
	TotalAmount    *jsonx.Int64String    `json:"totalAmount,omitempty"`
	BuyerPayAmount *jsonx.Int64String    `json:"buyerPayAmount,omitempty"`
	ReceiptAmount  *jsonx.Int64String    `json:"receiptAmount,omitempty"`
	BuyerUserID    string                `json:"buyerUserId,omitempty"`
	BuyerOpenID    string                `json:"buyerOpenId,omitempty"`
}

// SyncReq：查单
type SyncReq struct {
	OutTradeNo string                `json:"outTradeNo"`
	TradeNo    string                `json:"tradeNo"`
	Credential *alipay.SdkCredential `json:"credential"`
}

// SyncResp：查单响应
type SyncResp struct {
	TradeStatus    string                `json:"tradeStatus,omitempty"`
	Code           string                `json:"code,omitempty"`
	SubCode        string                `json:"subCode,omitempty"`
	SubMsg         string                `json:"subMsg,omitempty"`
	TradeNo        string                `json:"tradeNo,omitempty"`
	OutTradeNo     string                `json:"outTradeNo,omitempty"`
	SendPayDate    *jsonx.OffsetDateTime `json:"sendPayDate,omitempty"`
	BuyerUserID    string                `json:"buyerUserId,omitempty"`
	BuyerOpenID    string                `json:"buyerOpenId,omitempty"`
	BuyerPayAmount *jsonx.Int64String    `json:"buyerPayAmount,omitempty"`
}

// CloseReq：关单/撤销
type CloseReq struct {
	OutTradeNo string                `json:"outTradeNo"`
	TradeNo    string                `json:"tradeNo"`
	UseCancel  bool                  `json:"useCancel"`
	Credential *alipay.SdkCredential `json:"credential"`
}

// CloseResp：关单响应
type CloseResp struct {
	OutTradeNo string `json:"outTradeNo,omitempty"`
	TradeNo    string `json:"tradeNo,omitempty"`
	Code       string `json:"code,omitempty"`
	SubCode    string `json:"subCode,omitempty"`
	SubMsg     string `json:"subMsg,omitempty"`
}

// RefundReq：退款
type RefundReq struct {
	OutTradeNo   string                `json:"outTradeNo"`
	TradeNo      string                `json:"tradeNo"`
	OutRequestNo string                `json:"outRequestNo"`
	RefundAmount jsonx.Int64String     `json:"refundAmount"`
	Credential   *alipay.SdkCredential `json:"credential"`
}

// RefundResp：退款响应
type RefundResp struct {
	OutTradeNo   string                `json:"outTradeNo,omitempty"`
	TradeNo      string                `json:"tradeNo,omitempty"`
	OutRequestNo string                `json:"outRequestNo,omitempty"`
	FundChange   string                `json:"fundChange,omitempty"`
	Complete     bool                  `json:"complete"`
	FinishTime   *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
	RefundAmount *jsonx.Int64String    `json:"refundAmount,omitempty"`
	BuyerUserID  string                `json:"buyerUserId,omitempty"`
	BuyerOpenID  string                `json:"buyerOpenId,omitempty"`
}

// RefundSyncReq：退款查询
type RefundSyncReq struct {
	OutTradeNo   string                `json:"outTradeNo"`
	TradeNo      string                `json:"tradeNo"`
	OutRequestNo string                `json:"outRequestNo"`
	Credential   *alipay.SdkCredential `json:"credential"`
}

// RefundSyncResp：退款查询响应
type RefundSyncResp struct {
	RefundStatus string                `json:"refundStatus,omitempty"`
	Code         string                `json:"code,omitempty"`
	SubCode      string                `json:"subCode,omitempty"`
	SubMsg       string                `json:"subMsg,omitempty"`
	OutTradeNo   string                `json:"outTradeNo,omitempty"`
	TradeNo      string                `json:"tradeNo,omitempty"`
	OutRequestNo string                `json:"outRequestNo,omitempty"`
	FinishTime   *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
	RefundAmount *jsonx.Int64String    `json:"refundAmount,omitempty"`
}

// CallbackParseReq：回调验签解析
type CallbackParseReq struct {
	Credential *alipay.SdkCredential `json:"credential"`
	Params     map[string]string     `json:"params"`
}

// CallbackParseResp：回调解析结果
type CallbackParseResp struct {
	Success     bool                  `json:"success"`
	TradeType   string                `json:"tradeType,omitempty"`
	OutTradeNo  string                `json:"outTradeNo,omitempty"`
	TradeNo     string                `json:"tradeNo,omitempty"`
	OutRefundNo string                `json:"outRefundNo,omitempty"`
	TradeStatus string                `json:"tradeStatus,omitempty"`
	Amount      *jsonx.Int64String    `json:"amount,omitempty"`
	FinishTime  *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
}

// AppAuthTokenReq：换 app_auth_token
type AppAuthTokenReq struct {
	AuthCode   string                `json:"authCode"`
	Credential *alipay.SdkCredential `json:"credential"`
}

// AppAuthTokenResp：授权令牌响应
type AppAuthTokenResp struct {
	Code            string `json:"code,omitempty"`
	SubCode         string `json:"subCode,omitempty"`
	SubMsg          string `json:"subMsg,omitempty"`
	AppAuthToken    string `json:"appAuthToken,omitempty"`
	AppRefreshToken string `json:"appRefreshToken,omitempty"`
	AuthAppID       string `json:"authAppId,omitempty"`
	UserID          string `json:"userId,omitempty"`
	OpenID          string `json:"openId,omitempty"`
	ExpiresIn       string `json:"expiresIn,omitempty"`
	ReExpiresIn     string `json:"reExpiresIn,omitempty"`
}
