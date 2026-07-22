package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/ums"
	"daxpay.open/dax-pay-channel-one-go/internal/ums/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/ums/openapi"
)

// Pay：银联商务下单
func Pay(ctx context.Context, req *dto.PayReq) (*dto.PayResp, error) {
	_ = ctx
	if req.Method == "" {
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", "支付方式(method)不能为空")
	}
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}

	switch req.Method {
	case dto.PayMethodQRCode:
		return qrPay(client, req)
	case dto.PayMethodAlipayH5:
		return h5Pay(client, req, func(p map[string]any) (string, error) { return client.AlipayH5(p) })
	case dto.PayMethodWechatH5:
		return wechatH5Pay(client, req)
	case dto.PayMethodWechatCashier:
		return wechatCashierPay(client, req)
	case dto.PayMethodUnionJSAPI:
		return h5Pay(client, req, func(p map[string]any) (string, error) { return client.UnionH5(p) })
	default:
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", "unsupported method: "+string(req.Method))
	}
}

func qrPay(client *openapi.Client, req *dto.PayReq) (*dto.PayResp, error) {
	param := baseParam(req.Credential)
	param["instMid"] = dto.InstMidQR
	param["billNo"] = req.OutTradeNo
	param["billDate"] = openapi.TodayDate()
	param["totalAmount"] = int64(req.Amount)
	param["notifyUrl"] = req.NotifyURL
	if req.LimitCreditCard != nil && *req.LimitCreditCard {
		param["limitCreditCard"] = "true"
	}
	if req.Description != "" {
		param["goods"] = []map[string]any{{"goodsName": req.Description}}
	}
	result, err := client.QrPay(param)
	if err != nil {
		return nil, err
	}
	qr := mapStr(result, "billQRCode")
	if qr == "" {
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", "银联商务扫码支付未返回 billQRCode")
	}
	return &dto.PayResp{
		OutTradeNo:  req.OutTradeNo,
		PayBody:     qr,
		PayBodyType: dto.PayBodyTypeQRCode,
	}, nil
}

func buildH5Base(req *dto.PayReq) map[string]any {
	param := baseParam(req.Credential)
	param["instMid"] = dto.InstMidH5
	param["merOrderId"] = req.OutTradeNo
	param["totalAmount"] = int64(req.Amount)
	param["notifyUrl"] = req.NotifyURL
	if req.LimitCreditCard != nil && *req.LimitCreditCard {
		param["limitCreditCard"] = "true"
	}
	return param
}

func h5Pay(client *openapi.Client, req *dto.PayReq, call func(map[string]any) (string, error)) (*dto.PayResp, error) {
	url, err := call(buildH5Base(req))
	if err != nil {
		return nil, err
	}
	return &dto.PayResp{
		OutTradeNo:  req.OutTradeNo,
		PayBody:     url,
		PayBodyType: dto.PayBodyTypeLink,
	}, nil
}

func wechatH5Pay(client *openapi.Client, req *dto.PayReq) (*dto.PayResp, error) {
	param := buildH5Base(req)
	param["sceneType"] = "AND_WAP"
	param["merAppName"] = req.Description
	param["merAppId"] = req.ClientIP
	return h5Pay(client, req, func(p map[string]any) (string, error) {
		_ = p
		return client.WechatH5(param)
	})
}

func wechatCashierPay(client *openapi.Client, req *dto.PayReq) (*dto.PayResp, error) {
	if req.WxAppID == "" {
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", "微信收银台支付必填 wxAppId")
	}
	param := buildH5Base(req)
	param["subAppId"] = req.WxAppID
	return h5Pay(client, req, func(p map[string]any) (string, error) {
		_ = p
		return client.WechatH5ToMini(param)
	})
}
