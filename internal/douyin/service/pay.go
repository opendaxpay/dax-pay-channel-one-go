package service

import (
	"context"
	"encoding/json"
	"net/http"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin"
	"daxpay.open/dax-pay-channel-one-go/internal/douyin/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/douyin/openapi"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// Pay：抖音下单（QR / JSAPI / H5 / APP）
func Pay(ctx context.Context, req *dto.PayReq) (*dto.PayResp, error) {
	middleware.LoggerWithTrace(ctx).Info("douyin pay",
		"outTradeNo", req.OutTradeNo, "amount", req.Amount, "method", req.Method)

	if req.Method == "" {
		// 抖音: 支付方式不能为空
		return nil, douyin.NewValidateError("channel.error.douyinPayMethodNull")
	}

	client, err := newClient(req.Credential)
	if err != nil {
		return nil, wrapAPIErr("channel.error.douyinPayFailed", err)
	}
	resp := &dto.PayResp{OutTradeNo: req.OutTradeNo}

	switch req.Method {
	case dto.PayMethodQR:
		err = payNative(ctx, client, req, resp)
	case dto.PayMethodJSAPI:
		err = payJSAPI(ctx, client, req, resp)
	case dto.PayMethodH5:
		err = payH5(ctx, client, req, resp)
	case dto.PayMethodAPP:
		err = payAPP(ctx, client, req, resp)
	default:
		return nil, douyin.NewValidateError("channel.error.douyinPayMethodNull")
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func payNative(ctx context.Context, client *openapi.Client, req *dto.PayReq, resp *dto.PayResp) error {
	body := buildPayBase(req.Credential, req.OutTradeNo, req.Description, req.NotifyURL, req.ClientIP, int64(req.Amount), req.ExpiredTime)
	raw, err := client.Do(ctx, http.MethodPost, "/v1/trade/transactions/native", body)
	if err != nil {
		return wrapAPIErr("channel.error.douyinPayFailed", err)
	}
	var out struct {
		CodeURL string `json:"code_url"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return douyin.NewSDKError("channel.error.douyinPayFailed", err.Error())
	}
	if out.CodeURL == "" {
		return douyin.NewSDKError("channel.error.douyinPayFailed", "未返回二维码链接")
	}
	resp.PayBody = out.CodeURL
	resp.PayBodyType = dto.PayBodyTypeQRCode
	return nil
}

func payJSAPI(ctx context.Context, client *openapi.Client, req *dto.PayReq, resp *dto.PayResp) error {
	if req.OpenID == "" {
		// 抖音: JSAPI 必填 openId
		return douyin.NewValidateError("channel.error.douyinJsapiNoOpenId")
	}
	body := buildPayBase(req.Credential, req.OutTradeNo, req.Description, req.NotifyURL, req.ClientIP, int64(req.Amount), req.ExpiredTime)
	body["payer"] = map[string]any{"openid": req.OpenID}
	raw, err := client.Do(ctx, http.MethodPost, "/v1/trade/transactions/jsapi", body)
	if err != nil {
		return wrapAPIErr("channel.error.douyinPayFailed", err)
	}
	var out struct {
		PrepayID string `json:"prepay_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return douyin.NewSDKError("channel.error.douyinPayFailed", err.Error())
	}
	if out.PrepayID == "" {
		return douyin.NewSDKError("channel.error.douyinPayFailed", "未返回prepay_id")
	}
	payBody, err := openapi.BuildJSAPIPayBody(req.Credential.DouyinAppID, out.PrepayID, req.Credential.MerchantPrivateKey)
	if err != nil {
		return douyin.NewSDKError("channel.error.douyinPayFailed", err.Error())
	}
	resp.PayBody = payBody
	resp.PayBodyType = dto.PayBodyTypeJSAPI
	return nil
}

func payAPP(ctx context.Context, client *openapi.Client, req *dto.PayReq, resp *dto.PayResp) error {
	body := buildPayBase(req.Credential, req.OutTradeNo, req.Description, req.NotifyURL, req.ClientIP, int64(req.Amount), req.ExpiredTime)
	raw, err := client.Do(ctx, http.MethodPost, "/v1/trade/transactions/app", body)
	if err != nil {
		return wrapAPIErr("channel.error.douyinPayFailed", err)
	}
	var out struct {
		PrepayID string `json:"prepay_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return douyin.NewSDKError("channel.error.douyinPayFailed", err.Error())
	}
	if out.PrepayID == "" {
		return douyin.NewSDKError("channel.error.douyinPayFailed", "未返回prepay_id")
	}
	resp.PayBody = out.PrepayID
	resp.PayBodyType = dto.PayBodyTypeIdentifier
	return nil
}

func payH5(ctx context.Context, client *openapi.Client, req *dto.PayReq, resp *dto.PayResp) error {
	body := buildPayBase(req.Credential, req.OutTradeNo, req.Description, req.NotifyURL, req.ClientIP, int64(req.Amount), req.ExpiredTime)
	scene := map[string]any{
		"h5_info": map[string]any{"type": h5TypeWap},
	}
	if req.ClientIP != "" {
		scene["payer_client_ip"] = req.ClientIP
	}
	body["scene_info"] = scene
	raw, err := client.Do(ctx, http.MethodPost, "/v1/trade/transactions/h5", body)
	if err != nil {
		return wrapAPIErr("channel.error.douyinPayFailed", err)
	}
	var out struct {
		H5URL string `json:"h5_url"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return douyin.NewSDKError("channel.error.douyinPayFailed", err.Error())
	}
	if out.H5URL == "" {
		return douyin.NewSDKError("channel.error.douyinPayFailed", "未返回h5_url")
	}
	resp.PayBody = out.H5URL
	resp.PayBodyType = dto.PayBodyTypeLink
	return nil
}
