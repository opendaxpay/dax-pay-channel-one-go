package service

import (
	"context"
	"crypto/rsa"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/openapi"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// ParsePayCallback：支付回调验签解析
func ParsePayCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	return doParseCallback(ctx, req, false)
}

// ParseRefundCallback：退款回调验签解析
func ParseRefundCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	return doParseCallback(ctx, req, true)
}

func doParseCallback(ctx context.Context, req *dto.CallbackParseReq, refund bool) *dto.CallbackParseResp {
	tradeType := "PAY"
	if refund {
		tradeType = "REFUND"
	}
	resp := &dto.CallbackParseResp{TradeType: tradeType, Success: false}
	if req == nil || len(req.Params) == 0 {
		middleware.LoggerWithTrace(ctx).Error("alipay callback params empty", "refund", refund)
		return resp
	}
	if !verifyCallbackSign(req.Credential, req.Params) {
		middleware.LoggerWithTrace(ctx).Error("alipay callback verify failed", "refund", refund)
		return resp
	}
	resp.Success = true
	params := req.Params
	if refund {
		resp.OutTradeNo = params["out_request_no"]
		resp.OutRefundNo = params["trade_no"]
		if params["refund_status"] == "REFUND_SUCCESS" {
			resp.TradeStatus = "SUCCESS"
		} else {
			resp.TradeStatus = "FAIL"
		}
		resp.Amount = alipay.YuanToFenPtr(params["refund_amount"])
		resp.FinishTime = alipay.ParseCst(params["gmt_refund"])
	} else {
		resp.OutTradeNo = params["out_trade_no"]
		resp.TradeNo = params["trade_no"]
		st := params["trade_status"]
		if st == "TRADE_SUCCESS" || st == "TRADE_FINISHED" {
			resp.TradeStatus = "SUCCESS"
		} else {
			resp.TradeStatus = "FAIL"
		}
		resp.Amount = alipay.YuanToFenPtr(params["total_amount"])
		resp.FinishTime = alipay.ParseCst(params["gmt_payment"])
	}
	return resp
}

func verifyCallbackSign(cred *alipay.SdkCredential, params map[string]string) bool {
	if cred == nil {
		return false
	}
	var pub *rsa.PublicKey
	var err error
	if cred.IsCert() {
		pub, err = openapi.ExtractRSAPublicKey(cred.AlipayCert)
	} else {
		pub, err = openapi.ParsePublicKey(cred.AlipayPublicKey)
	}
	if err != nil || pub == nil {
		return false
	}
	return openapi.RsaCheckV1(params, pub) == nil
}
