package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/ums/dto"
)

// Close：关单
func Close(ctx context.Context, req *dto.CloseReq) (*dto.CloseResp, error) {
	_ = ctx
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	param := baseParam(req.Credential)
	if isQR(req.Method) {
		param["qrCodeId"] = req.QRCodeID
		param["instMid"] = dto.InstMidQR
		param["attachRefund"] = true
		if _, err := client.CloseQr(param); err != nil {
			return nil, err
		}
	} else {
		param["merOrderId"] = req.OutTradeNo
		// 按 Java 原样：H5 关单也用 QRPAYDEFAULT
		param["instMid"] = dto.InstMidQR
		if _, err := client.CloseH5(param); err != nil {
			return nil, err
		}
	}
	return &dto.CloseResp{OutTradeNo: req.OutTradeNo}, nil
}
