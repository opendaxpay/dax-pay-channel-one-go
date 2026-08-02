package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/union/dto"
)

// Close 云闪付关单(银联交易类型 31, 需原交易查询凭证 queryId)
func Close(ctx context.Context, req *dto.CloseReq) (*dto.CloseResp, error) {
	_ = ctx
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	param := baseParam(req.Credential)
	param["orderId"] = req.OutTradeNo
	param["origQryId"] = req.QueryID
	if _, err := client.CloseOrder(param); err != nil {
		return nil, err
	}
	return &dto.CloseResp{OutTradeNo: req.OutTradeNo}, nil
}
