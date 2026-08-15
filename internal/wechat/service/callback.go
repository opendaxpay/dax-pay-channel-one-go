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

// transferNotifyResult：微信转账回调解密后的业务数据(对齐 TransferBillsNotifyResult.DecryptNotifyResult)
type transferNotifyResult struct {
	MchID          string `json:"mch_id"`
	OutBillNo      string `json:"out_bill_no"`
	TransferBillNo string `json:"transfer_bill_no"`
	State          string `json:"state"`
	TransferAmount *int64 `json:"transfer_amount"`
	Openid         string `json:"openid"`
	CreateTime     string `json:"create_time"`
	UpdateTime     string `json:"update_time"`
	FailReason     string `json:"fail_reason"`
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

// ParseTransferCallback：转账回调验签 + AEAD 解密(商家转账到零钱 V3 异步通知)
func ParseTransferCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.TransferCallbackParseResp {
	fail := &dto.TransferCallbackParseResp{Verified: false}
	client, err := newCallbackClient(req.Credential)
	if err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat transfer callback: bad credential", "err", err)
		return fail
	}
	if err := client.VerifyNotify(req.Timestamp, req.Nonce, req.Body, req.Signature, req.Serial); err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat transfer callback verify failed", "err", err)
		return fail
	}
	var env notifyEnvelope
	if err := json.Unmarshal([]byte(req.Body), &env); err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat transfer callback parse envelope", "err", err)
		return fail
	}
	plain, err := client.DecryptResource(
		env.Resource.AssociatedData,
		env.Resource.Nonce,
		env.Resource.Ciphertext,
	)
	if err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat transfer callback decrypt", "err", err)
		return fail
	}
	var result transferNotifyResult
	if err := json.Unmarshal(plain, &result); err != nil {
		middleware.LoggerWithTrace(ctx).Error("wechat transfer callback parse plain", "err", err)
		return fail
	}
	// update_time 原串透传(主应用解析), 与 Java 行为一致
	return &dto.TransferCallbackParseResp{
		Verified:       true,
		OutBillNo:      result.OutBillNo,
		TransferBillNo: result.TransferBillNo,
		TransferState:  result.State,
		FailReason:     result.FailReason,
		UpdateTime:     result.UpdateTime,
	}
}
