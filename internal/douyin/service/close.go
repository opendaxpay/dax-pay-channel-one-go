package service

import (
	"context"
	"net/http"
	"net/url"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// Close：关闭订单（POST .../out-trade-no/{id}/close，body 仅 mchid）
func Close(ctx context.Context, req *dto.CloseReq) (*dto.CloseResp, error) {
	middleware.LoggerWithTrace(ctx).Info("douyin close", "outTradeNo", req.OutTradeNo)

	client, err := newClient(req.Credential)
	if err != nil {
		return nil, wrapAPIErr("channel.error.douyinCloseFailed", err)
	}
	path := "/v1/trade/transactions/out-trade-no/" + url.PathEscape(req.OutTradeNo) + "/close"
	body := map[string]any{"mchid": req.Credential.MchID}
	if _, err := client.Do(ctx, http.MethodPost, path, body); err != nil {
		return nil, wrapAPIErr("channel.error.douyinCloseFailed", err)
	}
	return &dto.CloseResp{OutTradeNo: req.OutTradeNo}, nil
}
