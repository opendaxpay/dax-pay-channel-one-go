package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/union/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/union/openapi"
)

// ParsePayCallback 云闪付支付回调验签解析
//
// 银联回调为 form 参数, 含 signature 与 signPubKeyCert(银联签名证书),
// 验签方式为 RSA2 证书公钥校验。按 txnType 自动分发支付(txnType=01)与退款(txnType=04)回调。
func ParsePayCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	_ = ctx
	if !openapi.VerifyCallback(req.Params, req.Credential) {
		return &dto.CallbackParseResp{Verified: false}
	}
	// 退款回调(txnType=04)也可能进入此入口, 按类型分发
	if req.Params["txnType"] == "04" {
		return parseRefundCallback(req.Params)
	}
	return parsePayCallback(req.Params)
}

// ParseRefundCallback 云闪付退款回调验签解析
func ParseRefundCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	_ = ctx
	if !openapi.VerifyCallback(req.Params, req.Credential) {
		return &dto.CallbackParseResp{Verified: false}
	}
	return parseRefundCallback(req.Params)
}

// parsePayCallback：解析支付回调(orderId/txnAmt/queryId/accNo)
func parsePayCallback(params map[string]string) *dto.CallbackParseResp {
	return &dto.CallbackParseResp{
		Verified:    true,
		TradeType:   "PAY",
		OutTradeNo:  params["orderId"],
		TradeStatus: mapCallbackStatus(params["respCode"]),
		Amount:      params["txnAmt"],
		RealAmount:  params["settleAmt"],
		FinishTime:  params["txnTime"],
		QueryID:     params["queryId"],
		BuyerID:     params["accNo"],
	}
}

// parseRefundCallback：解析退款回调(orderId=退款单号)
func parseRefundCallback(params map[string]string) *dto.CallbackParseResp {
	return &dto.CallbackParseResp{
		Verified:    true,
		TradeType:   "REFUND",
		OutRefundNo: params["orderId"],
		TradeStatus: mapCallbackStatus(params["respCode"]),
		Amount:      params["txnAmt"],
		FinishTime:  params["txnTime"],
	}
}

// mapCallbackStatus：respCode 00→SUCCESS, 其他→PROGRESS(等待重试或同步确认)
func mapCallbackStatus(respCode string) string {
	if respCode == "00" {
		return "SUCCESS"
	}
	return "PROGRESS"
}
