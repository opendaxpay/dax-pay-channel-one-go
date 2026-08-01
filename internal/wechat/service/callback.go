package service

import (
	"context"
	"encoding/json"

	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/dto"
)

// notifyEnvelope：微信 V3 回调外层
type notifyEnvelope struct {
	Resource struct {
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		Nonce          string `json:"nonce"`
	} `json:"resource"`
}

// ParsePayCallback：支付回调验签 + AEAD 解密
func ParsePayCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	fail := &dto.CallbackParseResp{Verified: false}
	client, err := newCallbackClient(req.Credential)
	if err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat pay callback: bad credential", "err", err)
		return fail
	}
	if err := client.VerifyNotify(req.Timestamp, req.Nonce, req.Body, req.Signature, req.Serial); err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat pay callback verify failed", "err", err)
		return fail
	}
	var env notifyEnvelope
	if err := json.Unmarshal([]byte(req.Body), &env); err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat pay callback parse envelope", "err", err)
		return fail
	}
	plain, err := client.DecryptResource(
		env.Resource.AssociatedData,
		env.Resource.Nonce,
		env.Resource.Ciphertext,
	)
	if err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat pay callback decrypt", "err", err)
		return fail
	}

	var result struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		SuccessTime   string `json:"success_time"`
		Amount        *struct {
			Total *int64 `json:"total"`
		} `json:"amount"`
		Payer *struct {
			Openid    string `json:"openid"`
			SpOpenid  string `json:"sp_openid"`
			SubOpenid string `json:"sub_openid"`
		} `json:"payer"`
	}
	if err := json.Unmarshal(plain, &result); err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat pay callback parse plain", "err", err)
		return fail
	}

	resp := &dto.CallbackParseResp{
		Verified:      true,
		TradeType:     "PAY",
		OutTradeNo:    result.OutTradeNo,
		TransactionID: result.TransactionID,
		TradeState:    result.TradeState,
		SuccessTime:   result.SuccessTime,
	}
	if result.Amount != nil && result.Amount.Total != nil {
		resp.Amount = ptrInt64(*result.Amount.Total)
	}
	if result.Payer != nil {
		switch {
		case result.Payer.Openid != "":
			resp.Openid = result.Payer.Openid
		case result.Payer.SubOpenid != "":
			resp.Openid = result.Payer.SubOpenid
		default:
			resp.Openid = result.Payer.SpOpenid
		}
	}
	return resp
}

// ParseRefundCallback：退款回调验签 + AEAD 解密
func ParseRefundCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	fail := &dto.CallbackParseResp{Verified: false}
	client, err := newCallbackClient(req.Credential)
	if err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat refund callback: bad credential", "err", err)
		return fail
	}
	if err := client.VerifyNotify(req.Timestamp, req.Nonce, req.Body, req.Signature, req.Serial); err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat refund callback verify failed", "err", err)
		return fail
	}
	var env notifyEnvelope
	if err := json.Unmarshal([]byte(req.Body), &env); err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat refund callback parse envelope", "err", err)
		return fail
	}
	plain, err := client.DecryptResource(
		env.Resource.AssociatedData,
		env.Resource.Nonce,
		env.Resource.Ciphertext,
	)
	if err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat refund callback decrypt", "err", err)
		return fail
	}

	var result struct {
		OutRefundNo  string `json:"out_refund_no"`
		RefundID     string `json:"refund_id"`
		RefundStatus string `json:"refund_status"`
		SuccessTime  string `json:"success_time"`
		Amount       *struct {
			Refund *int64 `json:"refund"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plain, &result); err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat refund callback parse plain", "err", err)
		return fail
	}

	resp := &dto.CallbackParseResp{
		Verified:     true,
		TradeType:    "REFUND",
		OutRefundNo:  result.OutRefundNo,
		RefundID:     result.RefundID,
		RefundStatus: result.RefundStatus,
		SuccessTime:  result.SuccessTime,
	}
	if result.Amount != nil && result.Amount.Refund != nil {
		resp.Amount = ptrInt64(*result.Amount.Refund)
	}
	return resp
}
