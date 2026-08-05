package service

import (
	"context"
	"encoding/json"
	"net/http"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin"
	"daxpay.open/dax-pay-channel-one-go/internal/douyin/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// Refund：发起退款
func Refund(ctx context.Context, req *dto.RefundReq) (*dto.RefundResp, error) {
	middleware.LoggerWithTrace(ctx).Info("douyin refund",
		"outTradeNo", req.OutTradeNo, "outRefundNo", req.OutRefundNo)

	client, err := newClient(req.Credential)
	if err != nil {
		return nil, wrapAPIErr("channel.error.douyinRefundFailed", err)
	}

	reason := req.Reason
	if reason == "" {
		reason = "退款"
	}
	body := map[string]any{
		"appid":         req.Credential.DouyinAppID,
		"mchid":         req.Credential.MchID,
		"out_trade_no":  req.OutTradeNo,
		"out_refund_no": req.OutRefundNo,
		"reason":        reason,
		"amount":        refundAmount(int64(req.RefundAmount), int64(req.TotalAmount)),
	}
	if req.NotifyURL != "" {
		body["notify_url"] = req.NotifyURL
	}

	raw, err := client.Do(ctx, http.MethodPost, "/v1/trade/refund/domestic/refunds", body)
	if err != nil {
		return nil, wrapAPIErr("channel.error.douyinRefundFailed", err)
	}

	var out struct {
		RefundID    string `json:"refund_id"`
		Status      string `json:"status"`
		SuccessTime string `json:"success_time"`
		OutRefundNo string `json:"out_refund_no"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, douyin.NewSDKError("channel.error.douyinRefundFailed", err.Error())
	}
	return &dto.RefundResp{
		OutRefundNo:  req.OutRefundNo,
		RefundID:     out.RefundID,
		RefundStatus: out.Status,
		FinishTime:   out.SuccessTime,
	}, nil
}
