package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/dto"
)

// DirectRefund：直连退款
func DirectRefund(ctx context.Context, req *dto.RefundReq) (*dto.RefundResp, error) {
	return refund(ctx, req, false)
}

// IsvRefund：服务商退款（body 带 sub_mchid）
func IsvRefund(ctx context.Context, req *dto.RefundReq) (*dto.RefundResp, error) {
	return refund(ctx, req, true)
}

func refund(ctx context.Context, req *dto.RefundReq, isv bool) (*dto.RefundResp, error) {
	middleware.LoggerWithTrace(ctx).Info("wechat refund",
		"outTradeNo", req.OutTradeNo, "outRefundNo", req.OutRefundNo, "isv", isv)

	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"out_refund_no": req.OutRefundNo,
		"amount": map[string]any{
			"refund":   int64(req.RefundAmount),
			"total":    int64(req.TotalAmount),
			"currency": currencyCNY,
		},
	}
	// 优先用 transactionId
	if req.TransactionID != "" {
		body["transaction_id"] = req.TransactionID
	} else {
		body["out_trade_no"] = req.OutTradeNo
	}
	if req.Reason != "" {
		body["reason"] = req.Reason
	}
	if req.NotifyURL != "" {
		body["notify_url"] = req.NotifyURL
	}
	if isv {
		body["sub_mchid"] = req.Credential.SubMchId
	}

	raw, err := client.Do(ctx, http.MethodPost, "/v3/refund/domestic/refunds", body)
	if err != nil {
		return nil, wrapAPIErr("channel.error.wechatRefundCallFailed", err)
	}
	return parseRefundResp(raw)
}

// DirectRefundSync：直连退款查询
func DirectRefundSync(ctx context.Context, req *dto.RefundSyncReq) (*dto.RefundSyncResp, error) {
	return refundSync(ctx, req, false)
}

// IsvRefundSync：服务商退款查询（query 带 sub_mchid）
func IsvRefundSync(ctx context.Context, req *dto.RefundSyncReq) (*dto.RefundSyncResp, error) {
	return refundSync(ctx, req, true)
}

func refundSync(ctx context.Context, req *dto.RefundSyncReq, isv bool) (*dto.RefundSyncResp, error) {
	middleware.LoggerWithTrace(ctx).Info("wechat refund sync", "outRefundNo", req.OutRefundNo, "isv", isv)

	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	path := "/v3/refund/domestic/refunds/" + url.PathEscape(req.OutRefundNo)
	if isv {
		q := url.Values{}
		q.Set("sub_mchid", req.Credential.SubMchId)
		path += "?" + q.Encode()
	}
	raw, err := client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, wrapAPIErr("channel.error.wechatRefundQueryFailed", err)
	}

	var out struct {
		Status        string `json:"status"`
		RefundID      string `json:"refund_id"`
		OutRefundNo   string `json:"out_refund_no"`
		TransactionID string `json:"transaction_id"`
		OutTradeNo    string `json:"out_trade_no"`
		SuccessTime   string `json:"success_time"`
		Amount        *struct {
			Refund      *int64 `json:"refund"`
			PayerRefund *int64 `json:"payer_refund"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, wechat.NewSDKError("channel.error.wechatRefundQueryFailed", err.Error())
	}
	resp := &dto.RefundSyncResp{
		Status:        out.Status,
		RefundID:      out.RefundID,
		OutRefundNo:   out.OutRefundNo,
		TransactionID: out.TransactionID,
		OutTradeNo:    out.OutTradeNo,
		FinishTime:    parseRFC3339(out.SuccessTime),
	}
	if out.Amount != nil {
		if out.Amount.Refund != nil {
			resp.RefundAmount = ptrInt64(*out.Amount.Refund)
		}
		if out.Amount.PayerRefund != nil {
			resp.PayerRefund = ptrInt64(*out.Amount.PayerRefund)
		}
	}
	return resp, nil
}

func parseRefundResp(raw []byte) (*dto.RefundResp, error) {
	var out struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		OutRefundNo   string `json:"out_refund_no"`
		RefundID      string `json:"refund_id"`
		Status        string `json:"status"`
		SuccessTime   string `json:"success_time"`
		Amount        *struct {
			Refund      *int64 `json:"refund"`
			PayerRefund *int64 `json:"payer_refund"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, wechat.NewSDKError("channel.error.wechatRefundCallFailed", err.Error())
	}
	resp := &dto.RefundResp{
		OutTradeNo:    out.OutTradeNo,
		TransactionID: out.TransactionID,
		OutRefundNo:   out.OutRefundNo,
		RefundID:      out.RefundID,
		Status:        out.Status,
		Complete:      out.Status == statusSuccess || out.Status == statusClosed,
		FinishTime:    parseRFC3339(out.SuccessTime),
	}
	if out.Amount != nil {
		if out.Amount.Refund != nil {
			resp.RefundAmount = ptrInt64(*out.Amount.Refund)
		}
		if out.Amount.PayerRefund != nil {
			resp.PayerRefund = ptrInt64(*out.Amount.PayerRefund)
		}
	}
	return resp, nil
}
