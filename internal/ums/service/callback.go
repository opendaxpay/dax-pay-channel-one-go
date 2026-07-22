package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/ums/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/ums/sdk"
)

// ParsePayCallback：支付回调验签解析
func ParsePayCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	_ = ctx
	resp := &dto.CallbackParseResp{}
	if req.Credential == nil || !sdk.VerifyCallback(req.Params, req.Credential.SecretKey) {
		resp.Verified = false
		return resp
	}
	resp.Verified = true
	resp.TradeType = "PAY"
	params := req.Params
	if _, ok := params["billNo"]; ok {
		parseQrPayCallback(params, resp)
	} else {
		parseH5PayCallback(params, resp)
	}
	return resp
}

// ParseRefundCallback：退款回调验签解析
func ParseRefundCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	_ = ctx
	resp := &dto.CallbackParseResp{}
	if req.Credential == nil || !sdk.VerifyCallback(req.Params, req.Credential.SecretKey) {
		resp.Verified = false
		return resp
	}
	resp.Verified = true
	resp.TradeType = "REFUND"
	params := req.Params
	resp.OutRefundNo = params["refundOrderId"]
	resp.TradeStatus = mapRefundCallbackStatus(params["status"])
	resp.Amount = params["refundAmount"]
	resp.FinishTime = params["refundPayTime"]
	resp.TargetSys = params["targetSys"]
	resp.TargetOrderID = params["refundTargetOrderId"]
	return resp
}

func parseQrPayCallback(params map[string]string, resp *dto.CallbackParseResp) {
	resp.OutTradeNo = params["billNo"]
	resp.TradeStatus = mapQrPayStatus(params["billStatus"])
	resp.Amount = params["totalAmount"]
	resp.RealAmount = params["receiptAmount"]
	if bp := params["billPayment"]; bp != "" {
		if m := asMap(bp); m != nil {
			resp.FinishTime = mapStr(m, "payTime")
			resp.BuyerID = mapStr(m, "buyerId")
			resp.TargetSys = mapStr(m, "targetSys")
			resp.TargetOrderID = mapStr(m, "targetOrderId")
		}
	}
}

func parseH5PayCallback(params map[string]string, resp *dto.CallbackParseResp) {
	resp.OutTradeNo = params["merOrderId"]
	resp.TradeStatus = mapH5PayStatus(params["status"])
	resp.Amount = params["totalAmount"]
	resp.RealAmount = params["receiptAmount"]
	resp.FinishTime = params["payTime"]
	resp.BuyerID = params["buyerId"]
	resp.TargetSys = params["targetSys"]
	resp.TargetOrderID = params["targetOrderId"]
}

func mapQrPayStatus(s string) string {
	switch s {
	case "PAID", "REFUND":
		return "SUCCESS"
	case "CLOSED":
		return "CLOSED"
	default:
		return "PROGRESS"
	}
}

func mapH5PayStatus(s string) string {
	switch s {
	case "TRADE_SUCCESS":
		return "SUCCESS"
	case "TRADE_CLOSED":
		return "CLOSED"
	default:
		return "PROGRESS"
	}
}

func mapRefundCallbackStatus(s string) string {
	switch s {
	case "TRADE_REFUND":
		return "SUCCESS"
	case "TRADE_CLOSED":
		return "CLOSED"
	default:
		return "PROGRESS"
	}
}
