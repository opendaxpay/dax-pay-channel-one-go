package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// Sync：支付查单；ORDER_NOT_EXIST → tradeState=CLOSED
func Sync(ctx context.Context, req *dto.SyncReq) (*dto.SyncResp, error) {
	middleware.LoggerWithTrace(ctx).Info("douyin sync", "outTradeNo", req.OutTradeNo)

	client, err := newClient(req.Credential)
	if err != nil {
		return &dto.SyncResp{OutTradeNo: req.OutTradeNo, ErrorMsg: err.Error()}, nil
	}

	q := url.Values{}
	q.Set("mchid", req.Credential.MchID)
	path := "/v1/trade/transactions/out-trade-no/" + url.PathEscape(req.OutTradeNo) + "?" + q.Encode()

	raw, err := client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		resp := &dto.SyncResp{OutTradeNo: req.OutTradeNo}
		if containsOrderNotExist(err) {
			resp.TradeState = tradeStateClosed
			return resp, nil
		}
		resp.ErrorMsg = err.Error()
		return resp, nil
	}

	var out struct {
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		SuccessTime   string `json:"success_time"`
		Amount        *struct {
			Total *int64 `json:"total"`
		} `json:"amount"`
		Payer *struct {
			Openid string `json:"openid"`
		} `json:"payer"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return &dto.SyncResp{OutTradeNo: req.OutTradeNo, ErrorMsg: err.Error()}, nil
	}

	resp := &dto.SyncResp{
		OutTradeNo:    req.OutTradeNo,
		TransactionID: out.TransactionID,
		TradeState:    out.TradeState,
		SuccessTime:   out.SuccessTime,
	}
	if out.Amount != nil && out.Amount.Total != nil {
		resp.TotalAmount = ptrInt64(*out.Amount.Total)
	}
	if out.Payer != nil {
		resp.OpenID = out.Payer.Openid
	}
	return resp, nil
}

// RefundSync：退款查单；异常不抛，写 errorMsg
func RefundSync(ctx context.Context, req *dto.RefundSyncReq) (*dto.RefundSyncResp, error) {
	middleware.LoggerWithTrace(ctx).Info("douyin refund-sync", "outRefundNo", req.OutRefundNo)

	client, err := newClient(req.Credential)
	if err != nil {
		return &dto.RefundSyncResp{OutRefundNo: req.OutRefundNo, ErrorMsg: err.Error()}, nil
	}

	q := url.Values{}
	q.Set("mchid", req.Credential.MchID)
	path := "/v1/trade/refund/domestic/refunds/" + url.PathEscape(req.OutRefundNo) + "?" + q.Encode()

	raw, err := client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return &dto.RefundSyncResp{OutRefundNo: req.OutRefundNo, ErrorMsg: err.Error()}, nil
	}

	var out struct {
		RefundID    string `json:"refund_id"`
		Status      string `json:"status"`
		SuccessTime string `json:"success_time"`
		Amount      *struct {
			Refund *int64 `json:"refund"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return &dto.RefundSyncResp{OutRefundNo: req.OutRefundNo, ErrorMsg: err.Error()}, nil
	}

	resp := &dto.RefundSyncResp{
		OutRefundNo:  req.OutRefundNo,
		RefundID:     out.RefundID,
		RefundStatus: out.Status,
		FinishTime:   out.SuccessTime,
	}
	if out.Amount != nil && out.Amount.Refund != nil {
		resp.RefundAmount = ptrInt64(*out.Amount.Refund)
	}
	return resp, nil
}
