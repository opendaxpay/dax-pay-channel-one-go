package service

import (
	"context"
	"encoding/json"
	"net/http"

	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/openapi"
)

// DirectPay：直连下单
func DirectPay(ctx context.Context, req *dto.PayReq) (*dto.PayResp, error) {
	return pay(ctx, req, false)
}

// IsvPay：服务商下单
func IsvPay(ctx context.Context, req *dto.PayReq) (*dto.PayResp, error) {
	return pay(ctx, req, true)
}

func pay(ctx context.Context, req *dto.PayReq, isv bool) (*dto.PayResp, error) {
	middleware.LoggerWithTrace(ctx).Info("wechat pay",
		"outTradeNo", req.OutTradeNo, "amount", req.Amount, "method", req.Method, "isv", isv)

	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	resp := &dto.PayResp{OutTradeNo: req.OutTradeNo, Complete: false}

	switch req.Method {
	case dto.PayMethodNative:
		err = payNative(ctx, client, req, resp, isv)
	case dto.PayMethodJSAPI, dto.PayMethodMini:
		err = payJSAPI(ctx, client, req, resp, isv)
	case dto.PayMethodApp:
		err = payApp(ctx, client, req, resp, isv)
	case dto.PayMethodH5:
		err = payH5(ctx, client, req, resp, isv)
	case dto.PayMethodMicropay:
		err = payCodepay(ctx, client, req, resp, isv)
	default:
		return nil, wechat.NewValidateError("channel.error.validateParams")
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func payPath(method string, isv bool) string {
	if isv {
		return "/v3/pay/partner/transactions/" + method
	}
	return "/v3/pay/transactions/" + method
}

func buildBaseBody(req *dto.PayReq, isv bool) map[string]any {
	cred := req.Credential
	body := map[string]any{
		"description":  req.Description,
		"out_trade_no": req.OutTradeNo,
		"amount":       amountMap(int64(req.Amount)),
	}
	if isv {
		body["sp_appid"] = cred.WxAppId
		body["sp_mchid"] = cred.WxMchId
		body["sub_mchid"] = cred.SubMchId
		if cred.SubAppId != "" {
			body["sub_appid"] = cred.SubAppId
		}
	} else {
		body["appid"] = cred.WxAppId
		body["mchid"] = cred.WxMchId
	}
	if req.Attach != "" {
		body["attach"] = req.Attach
	}
	if req.NotifyURL != "" {
		body["notify_url"] = req.NotifyURL
	}
	if req.ExpireTime != nil {
		body["time_expire"] = formatExpire(*req.ExpireTime)
	}
	// 分账订单: 透传 settle_info.profit_sharing=true
	if req.Allocation != nil && *req.Allocation {
		body["settle_info"] = map[string]any{"profit_sharing": true}
	}
	return body
}

func payNative(ctx context.Context, client *openapi.Client, req *dto.PayReq, resp *dto.PayResp, isv bool) error {
	body := buildBaseBody(req, isv)
	raw, err := client.Do(ctx, http.MethodPost, payPath("native", isv), body)
	if err != nil {
		return wrapAPIErr("channel.error.wechatPayCallFailed", err)
	}
	var out struct {
		CodeURL string `json:"code_url"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return wechat.NewSDKError("channel.error.wechatPayCallFailed", err.Error())
	}
	resp.PayBody = out.CodeURL
	resp.PayBodyType = dto.PayBodyQRCode
	return nil
}

func payJSAPI(ctx context.Context, client *openapi.Client, req *dto.PayReq, resp *dto.PayResp, isv bool) error {
	if req.OpenID == "" {
		// 微信: JSAPI/小程序支付必填 openid
		return wechat.NewValidateError("channel.error.wechatOpenIdRequired")
	}
	body := buildBaseBody(req, isv)
	if isv {
		payer := map[string]any{}
		if req.Credential.SubAppId != "" {
			body["sub_appid"] = req.Credential.SubAppId
			payer["sub_openid"] = req.OpenID
		} else {
			payer["sp_openid"] = req.OpenID
		}
		body["payer"] = payer
	} else {
		body["payer"] = map[string]any{"openid": req.OpenID}
	}
	raw, err := client.Do(ctx, http.MethodPost, payPath("jsapi", isv), body)
	if err != nil {
		return wrapAPIErr("channel.error.wechatPayCallFailed", err)
	}
	var out struct {
		PrepayID string `json:"prepay_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return wechat.NewSDKError("channel.error.wechatPayCallFailed", err.Error())
	}
	// JSAPI 调起 appId：ISV 有 sub_appid 时用子应用，否则服务商/直连 appId
	appID := req.Credential.WxAppId
	if isv && req.Credential.SubAppId != "" {
		appID = req.Credential.SubAppId
	}
	payBody, err := openapi.SignJSAPIPayBody(appID, out.PrepayID, client.SignMessage)
	if err != nil {
		return wechat.NewSDKError("channel.error.wechatPayCallFailed", err.Error())
	}
	resp.PayBody = payBody
	resp.PayBodyType = dto.PayBodyJSAPI
	return nil
}

func payApp(ctx context.Context, client *openapi.Client, req *dto.PayReq, resp *dto.PayResp, isv bool) error {
	body := buildBaseBody(req, isv)
	raw, err := client.Do(ctx, http.MethodPost, payPath("app", isv), body)
	if err != nil {
		return wrapAPIErr("channel.error.wechatPayCallFailed", err)
	}
	var out struct {
		PrepayID string `json:"prepay_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return wechat.NewSDKError("channel.error.wechatPayCallFailed", err.Error())
	}
	appID := req.Credential.WxAppId
	if isv && req.Credential.SubAppId != "" {
		appID = req.Credential.SubAppId
	}
	payBody, err := openapi.SignAppPayBody(appID, req.Credential.WxMchId, out.PrepayID, client.SignMessage)
	if err != nil {
		return wechat.NewSDKError("channel.error.wechatPayCallFailed", err.Error())
	}
	resp.PayBody = payBody
	resp.PayBodyType = dto.PayBodyAppOrderStr
	return nil
}

func payH5(ctx context.Context, client *openapi.Client, req *dto.PayReq, resp *dto.PayResp, isv bool) error {
	if req.PayerClientIp == "" || req.WapURL == "" {
		// 微信: H5 支付必填 payerClientIp 与 wapUrl
		return wechat.NewValidateError("channel.error.wechatH5SceneRequired")
	}
	body := buildBaseBody(req, isv)
	h5Info := map[string]any{
		"type":    h5SceneType,
		"app_url": req.WapURL,
	}
	if req.WapName != "" {
		h5Info["app_name"] = req.WapName
	}
	body["scene_info"] = map[string]any{
		"payer_client_ip": req.PayerClientIp,
		"h5_info":         h5Info,
	}
	raw, err := client.Do(ctx, http.MethodPost, payPath("h5", isv), body)
	if err != nil {
		return wrapAPIErr("channel.error.wechatPayCallFailed", err)
	}
	var out struct {
		H5URL string `json:"h5_url"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return wechat.NewSDKError("channel.error.wechatPayCallFailed", err.Error())
	}
	resp.PayBody = out.H5URL
	resp.PayBodyType = dto.PayBodyLink
	return nil
}

func payCodepay(ctx context.Context, client *openapi.Client, req *dto.PayReq, resp *dto.PayResp, isv bool) error {
	if req.AuthCode == "" {
		// 微信: 付款码支付必填 authCode
		return wechat.NewValidateError("channel.error.wechatAuthCodeRequired")
	}
	cred := req.Credential
	body := map[string]any{
		"description":  req.Description,
		"out_trade_no": req.OutTradeNo,
		"amount":       amountMap(int64(req.Amount)),
		"scene_info": map[string]any{
			"store_info": map[string]any{"out_id": "1"},
		},
		"payer": map[string]any{"auth_code": req.AuthCode},
	}
	if isv {
		body["sp_appid"] = cred.WxAppId
		body["sp_mchid"] = cred.WxMchId
		body["sub_mchid"] = cred.SubMchId
	} else {
		body["appid"] = cred.WxAppId
		body["mchid"] = cred.WxMchId
	}
	if req.Attach != "" {
		body["attach"] = req.Attach
	}
	// 分账订单: 透传 settle_info.profit_sharing=true
	if req.Allocation != nil && *req.Allocation {
		body["settle_info"] = map[string]any{"profit_sharing": true}
	}

	raw, err := client.Do(ctx, http.MethodPost, payPath("codepay", isv), body)
	if err != nil {
		if isCodepayPaying(err) {
			// 微信: 付款码用户支付中, 需主应用轮询同步
			return wechat.NewResultUnknown("channel.error.wechatCodepayUserPaying", err.Error())
		}
		return wrapAPIErr("channel.error.wechatPayCallFailed", err)
	}

	var out struct {
		TransactionID string `json:"transaction_id"`
		SuccessTime   string `json:"success_time"`
		Amount        *struct {
			Total      *int64 `json:"total"`
			PayerTotal *int64 `json:"payer_total"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return wechat.NewSDKError("channel.error.wechatPayCallFailed", err.Error())
	}
	resp.TransactionID = out.TransactionID
	resp.Complete = true
	resp.FinishTime = parseRFC3339(out.SuccessTime)
	if out.Amount != nil {
		if out.Amount.Total != nil {
			resp.TotalAmount = ptrInt64(*out.Amount.Total)
		}
		if out.Amount.PayerTotal != nil {
			resp.PayerTotal = ptrInt64(*out.Amount.PayerTotal)
		}
	}
	return nil
}
