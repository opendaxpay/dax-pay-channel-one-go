package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
)

// ParsePayCallback：支付回调验签解密
func ParsePayCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	return parseCallback(ctx, req, "PAY")
}

// ParseRefundCallback：退款回调验签解密
func ParseRefundCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	return parseCallback(ctx, req, "REFUND")
}

func parseCallback(ctx context.Context, req *dto.CallbackParseReq, tradeType string) *dto.CallbackParseResp {
	resp := &dto.CallbackParseResp{TradeType: tradeType, Verified: false}
	if req.Credential == nil {
		return resp
	}
	client, err := newClient(req.Credential)
	if err != nil {
		return resp
	}
	if err := client.VerifyNotify(ctx, req.Timestamp, req.Nonce, req.Body, req.Signature, req.Serial); err != nil {
		return resp
	}
	var envelope struct {
		Resource struct {
			Ciphertext     string `json:"ciphertext"`
			AssociatedData string `json:"associated_data"`
			Nonce          string `json:"nonce"`
		} `json:"resource"`
	}
	if json.Unmarshal([]byte(req.Body), &envelope) != nil {
		return resp
	}
	plain, err := client.DecryptResource(envelope.Resource.AssociatedData, envelope.Resource.Nonce, envelope.Resource.Ciphertext)
	if err != nil {
		return resp
	}
	var data map[string]any
	if json.Unmarshal(plain, &data) != nil {
		return resp
	}
	resp.Verified = true
	if tradeType == "PAY" {
		resp.OutTradeNo = anyString(data["out_trade_no"])
		resp.TransactionID = anyString(data["transaction_id"])
		resp.TradeState = anyString(data["trade_state"])
		resp.SuccessTime = anyString(data["success_time"])
		if amount, ok := data["amount"].(map[string]any); ok {
			resp.Amount = anyInt64Ptr(amount["total"])
		}
		if payer, ok := data["payer"].(map[string]any); ok {
			resp.OpenID = anyString(payer["openid"])
		}
	} else {
		resp.OutTradeNo = anyString(data["out_trade_no"])
		resp.OutRefundNo = anyString(data["out_refund_no"])
		resp.RefundID = anyString(data["refund_id"])
		resp.RefundStatus = anyString(data["refund_status"])
		resp.SuccessTime = anyString(data["success_time"])
		if amount, ok := data["amount"].(map[string]any); ok {
			resp.Amount = anyInt64Ptr(amount["refund"])
		}
	}
	return resp
}

func anyString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func anyInt64Ptr(v any) *jsonx.Int64String {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case float64:
		return ptrInt64(int64(t))
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil
		}
		return ptrInt64(n)
	default:
		return nil
	}
}
