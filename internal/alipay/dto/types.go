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
	ReturnURL  string                `json:"returnUrl"`
	AuthCode   string                `json:"authCode"`
	OpenID     string                `json:"openId"`
	Allocation *bool                 `json:"allocation"`
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

// ReportInfo：转账场景报备信息(transfer_scene_report_infos)
type ReportInfo struct {
	InfoType    string `json:"infoType"`
	InfoContent string `json:"infoContent"`
}

// TransferReq：转账发起/同步共用请求(对标 AlipayTransferReq)
//
// 发起时 amount/payeeType/payeeAccount 必填; 同步按 outBizNo 反查, 其余字段可空。
type TransferReq struct {
	OutBizNo          string                `json:"outBizNo"`
	Amount            jsonx.Int64String     `json:"amount"`
	Title             string                `json:"title"`
	Remark            string                `json:"remark"`
	PayeeType         string                `json:"payeeType"`
	PayeeAccount      string                `json:"payeeAccount"`
	PayeeName         string                `json:"payeeName"`
	NotifyURL         string                `json:"notifyUrl"`
	TransferSceneName string                `json:"transferSceneName"`
	ReportInfos       []ReportInfo          `json:"reportInfos"`
	Credential        *alipay.SdkCredential `json:"credential"`
}

// TransferResp：转账响应(对标 AlipayTransferResp)
type TransferResp struct {
	OrderID        string `json:"orderId,omitempty"`
	Status         string `json:"status,omitempty"`
	FailReason     string `json:"failReason,omitempty"`
	FinishTime     string `json:"finishTime,omitempty"`
	PayFundOrderID string `json:"payFundOrderId,omitempty"`
	TransDate      string `json:"transDate,omitempty"`
	ErrorCode      string `json:"errorCode,omitempty"`
	Code           string `json:"code,omitempty"`
	SubCode        string `json:"subCode,omitempty"`
	SubMsg         string `json:"subMsg,omitempty"`
}

// RoyaltyParam：分账子参数
type RoyaltyParam struct {
	TransInType string            `json:"transInType"`
	TransIn     string            `json:"transIn"`
	Amount      jsonx.Int64String `json:"amount"`
}

// AllocReq：分账发起/同步共用请求(对标 AlipayAllocReq)
type AllocReq struct {
	OutRequestNo     string                `json:"outRequestNo"`
	TradeNo          string                `json:"tradeNo"`
	RoyaltyMode      string                `json:"royaltyMode"`
	RoyaltyParameters []RoyaltyParam       `json:"royaltyParameters"`
	Credential       *alipay.SdkCredential `json:"credential"`
}

// RoyaltyDetailResult：分账逐明细结果
type RoyaltyDetailResult struct {
	DetailID  string `json:"detailId,omitempty"`
	TransIn   string `json:"transIn,omitempty"`
	State     string `json:"state,omitempty"`
	ErrorDesc string `json:"errorDesc,omitempty"`
	ExecuteDt string `json:"executeDt,omitempty"`
}

// AllocResp：分账响应(对标 AlipayAllocResp)
type AllocResp struct {
	SettleNo          string                `json:"settleNo,omitempty"`
	RoyaltyDetailList []RoyaltyDetailResult `json:"royaltyDetailList,omitempty"`
	Code              string                `json:"code,omitempty"`
	SubCode           string                `json:"subCode,omitempty"`
	SubMsg            string                `json:"subMsg,omitempty"`
}

// TransferCallbackParseResp：转账回调解析响应(对标 AlipayTransferCallbackParseResp)
type TransferCallbackParseResp struct {
	Success        bool                  `json:"success"`
	OutBizNo       string                `json:"outBizNo,omitempty"`
	OrderID        string                `json:"orderId,omitempty"`
	TransferStatus string                `json:"transferStatus,omitempty"`
	FailReason     string                `json:"failReason,omitempty"`
	FinishTime     *jsonx.OffsetDateTime `json:"finishTime,omitempty"`
}
