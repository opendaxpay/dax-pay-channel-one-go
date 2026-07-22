package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"daxpay.open/dax-pay-channel-one-go/internal/wechat"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/dto"
)

// DirectSync：直连查单
func DirectSync(ctx context.Context, req *dto.SyncReq) (*dto.SyncResp, error) {
	return syncOrder(ctx, req, false)
}

// IsvSync：服务商查单
func IsvSync(ctx context.Context, req *dto.SyncReq) (*dto.SyncResp, error) {
	return syncOrder(ctx, req, true)
}

func syncOrder(ctx context.Context, req *dto.SyncReq, isv bool) (*dto.SyncResp, error) {
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	cred := req.Credential
	var path string
	if req.TransactionID != "" {
		// 优先用 transactionId 查询
		if isv {
			q := url.Values{}
			q.Set("sp_mchid", cred.WxMchId)
			q.Set("sub_mchid", cred.SubMchId)
			path = "/v3/pay/partner/transactions/id/" + url.PathEscape(req.TransactionID) + "?" + q.Encode()
		} else {
			q := url.Values{}
			q.Set("mchid", cred.WxMchId)
			path = "/v3/pay/transactions/id/" + url.PathEscape(req.TransactionID) + "?" + q.Encode()
		}
	} else {
		if isv {
			q := url.Values{}
			q.Set("sp_mchid", cred.WxMchId)
			q.Set("sub_mchid", cred.SubMchId)
			path = "/v3/pay/partner/transactions/out-trade-no/" + url.PathEscape(req.OutTradeNo) + "?" + q.Encode()
		} else {
			q := url.Values{}
			q.Set("mchid", cred.WxMchId)
			path = "/v3/pay/transactions/out-trade-no/" + url.PathEscape(req.OutTradeNo) + "?" + q.Encode()
		}
	}

	raw, err := client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, wrapAPIErr("channel.error.wechatOrderQueryFailed", err)
	}

	var out struct {
		TradeState     string `json:"trade_state"`
		TradeStateDesc string `json:"trade_state_desc"`
		TransactionID  string `json:"transaction_id"`
		OutTradeNo     string `json:"out_trade_no"`
		SuccessTime    string `json:"success_time"`
		Amount         *struct {
			Total      *int64 `json:"total"`
			PayerTotal *int64 `json:"payer_total"`
		} `json:"amount"`
		Payer *struct {
			Openid    string `json:"openid"`
			SpOpenid  string `json:"sp_openid"`
			SubOpenid string `json:"sub_openid"`
		} `json:"payer"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, wechat.NewSDKError("channel.error.wechatOrderQueryFailed", err.Error())
	}

	resp := &dto.SyncResp{
		TradeState:     out.TradeState,
		TradeStateDesc: out.TradeStateDesc,
		TransactionID:  out.TransactionID,
		OutTradeNo:     out.OutTradeNo,
		SuccessTime:    parseRFC3339(out.SuccessTime),
	}
	if out.Amount != nil {
		if out.Amount.Total != nil {
			resp.TotalAmount = ptrInt64(*out.Amount.Total)
		}
		if out.Amount.PayerTotal != nil {
			resp.PayerTotal = ptrInt64(*out.Amount.PayerTotal)
		}
	}
	if out.Payer != nil {
		if isv {
			// sub_openid 优先, 空则回退 sp_openid
			if out.Payer.SubOpenid != "" {
				resp.OpenID = out.Payer.SubOpenid
			} else {
				resp.OpenID = out.Payer.SpOpenid
			}
		} else {
			resp.OpenID = out.Payer.Openid
		}
	}
	return resp, nil
}
