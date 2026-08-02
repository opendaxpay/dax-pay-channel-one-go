package service

import (
	"context"
	"fmt"

	"daxpay.open/dax-pay-channel-one-go/internal/union"
	"daxpay.open/dax-pay-channel-one-go/internal/union/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/union/openapi"
)

// Pay 云闪付下单
func Pay(ctx context.Context, req *dto.PayReq) (*dto.PayResp, error) {
	_ = ctx
	if req.Method == "" {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", "支付方式(method)不能为空")
	}
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	switch req.Method {
	case dto.PayMethodQRCode:
		return applyQRCode(client, req)
	case dto.PayMethodBarcode:
		return consume(client, req)
	case dto.PayMethodH5:
		return wapPay(client, req)
	default:
		return nil, union.NewSDKError("channel.error.unionRequestFailed", "unsupported method: "+string(req.Method))
	}
}

// applyQRCode：主扫支付(申请二维码, 返回 qrNo 二维码内容)
func applyQRCode(client *openapi.Client, req *dto.PayReq) (*dto.PayResp, error) {
	param := baseParam(req.Credential)
	param["orderId"] = req.OutTradeNo
	param["txnAmt"] = fmt.Sprintf("%d", int64(req.Amount))
	if req.Description != "" {
		param["orderDesc"] = req.Description
	}
	param["backUrl"] = req.NotifyURL
	result, err := client.ApplyQRCode(param)
	if err != nil {
		return nil, err
	}
	qr := result["qrNo"]
	if qr == "" {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", "未返回 qrNo")
	}
	return &dto.PayResp{
		OutTradeNo:  req.OutTradeNo,
		PayBody:     qr,
		PayBodyType: dto.PayBodyTypeQRCode,
	}, nil
}

// consume：被扫支付(付款码消费, 同步返回支付结果)
//
// respCode=00 支付成功 / 03 处理中, 其他视为失败。被扫无支付内容, 主应用走 sync/回调确认。
func consume(client *openapi.Client, req *dto.PayReq) (*dto.PayResp, error) {
	if req.AuthCode == "" {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", "被扫支付必填 authCode")
	}
	param := baseParam(req.Credential)
	param["orderId"] = req.OutTradeNo
	param["txnAmt"] = fmt.Sprintf("%d", int64(req.Amount))
	param["qrNo"] = req.AuthCode
	result, err := client.Consume(param)
	if err != nil {
		return nil, err
	}
	respCode := result["respCode"]
	if respCode != "00" && respCode != "03" {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", result["respMsg"])
	}
	return &dto.PayResp{OutTradeNo: req.OutTradeNo}, nil
}

// wapPay：H5/WAP 网关支付(返回自动提交 HTML form)
func wapPay(client *openapi.Client, req *dto.PayReq) (*dto.PayResp, error) {
	param := baseParam(req.Credential)
	param["orderId"] = req.OutTradeNo
	param["txnAmt"] = fmt.Sprintf("%d", int64(req.Amount))
	if req.Description != "" {
		param["orderDesc"] = req.Description
	}
	param["backUrl"] = req.NotifyURL
	html, err := client.BuildWapFormHTML(param)
	if err != nil {
		return nil, err
	}
	return &dto.PayResp{
		OutTradeNo:  req.OutTradeNo,
		PayBody:     html,
		PayBodyType: dto.PayBodyTypeLink,
	}, nil
}
