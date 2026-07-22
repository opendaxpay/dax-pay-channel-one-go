package service

import (
	"context"
	"net/http"
	"net/url"

	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/dto"
)

// DirectClose：直连关单（ORDER_NOT_EXIST / ORDER_CLOSED 视为成功）
func DirectClose(ctx context.Context, req *dto.CloseReq) (*dto.CloseResp, error) {
	return closeOrder(ctx, req, false)
}

// IsvClose：服务商关单
func IsvClose(ctx context.Context, req *dto.CloseReq) (*dto.CloseResp, error) {
	return closeOrder(ctx, req, true)
}

func closeOrder(ctx context.Context, req *dto.CloseReq, isv bool) (*dto.CloseResp, error) {
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	cred := req.Credential
	escaped := url.PathEscape(req.OutTradeNo)
	var path string
	var body map[string]any
	if isv {
		path = "/v3/pay/partner/transactions/out-trade-no/" + escaped + "/close"
		body = map[string]any{
			"sp_mchid":  cred.WxMchId,
			"sub_mchid": cred.SubMchId,
		}
	} else {
		path = "/v3/pay/transactions/out-trade-no/" + escaped + "/close"
		body = map[string]any{"mchid": cred.WxMchId}
	}

	_, err = client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		if isCloseSuccessFallback(err) {
			middleware.LoggerWithTrace(ctx).Info("wechat close fallback ok",
				"outTradeNo", req.OutTradeNo, "err", err.Error())
			return &dto.CloseResp{OutTradeNo: req.OutTradeNo, TransactionID: req.TransactionID}, nil
		}
		return nil, wrapAPIErr("channel.error.wechatCloseOrderFailed", err)
	}
	return &dto.CloseResp{OutTradeNo: req.OutTradeNo, TransactionID: req.TransactionID}, nil
}
