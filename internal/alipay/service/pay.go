package service

import (
	"context"
	"encoding/json"
	"strings"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/openapi"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

const (
	codeSuccess   = "10000"
	codeInProcess = "10003"
	buyerIDPrefix = "2088"
)

// gatewayBiz：网关业务响应公共字段
type gatewayBiz struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	SubCode string `json:"sub_code"`
	SubMsg  string `json:"sub_msg"`
}

func (g gatewayBiz) success() bool {
	return g.Code == codeSuccess
}

func (g gatewayBiz) errDetail() string {
	if g.SubMsg != "" {
		return g.SubMsg
	}
	return g.Msg
}

func newClient(cred *alipay.SdkCredential) (*openapi.Client, error) {
	return openapi.NewClient(cred)
}

func executeJSON(ctx context.Context, cred *alipay.SdkCredential, method string, biz any, notifyURL string, out any) error {
	client, err := newClient(cred)
	if err != nil {
		return alipay.NewSDKError("channel.error.sdkCallFailedWithDetail", err.Error())
	}
	raw, err := client.Execute(ctx, method, biz, notifyURL)
	if err != nil {
		return alipay.NewSDKError("channel.error.sdkCallFailedWithDetail", err.Error())
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return alipay.NewSDKError("channel.error.sdkCallFailedWithDetail", err.Error())
	}
	return nil
}

// Pay：支付宝下单（六种 method）
func Pay(ctx context.Context, req *dto.PayReq) (*dto.PayResp, error) {
	middleware.LoggerWithTrace(ctx).Info("alipay pay",
		"outTradeNo", req.OutTradeNo, "amount", req.Amount, "method", req.Method)

	amount := alipay.FenToYuan(int64(req.Amount))
	resp := &dto.PayResp{OutTradeNo: req.OutTradeNo, Complete: false}

	client, err := newClient(req.Credential)
	if err != nil {
		return nil, alipay.NewSDKError("channel.error.alipayPayCallFailed", err.Error())
	}

	var expire string
	if req.ExpireTime != nil {
		expire = alipay.FormatExpire(*req.ExpireTime)
	}

	switch req.Method {
	case dto.MethodWAP:
		err = payPage(client, "alipay.trade.wap.pay", map[string]any{
			"out_trade_no": req.OutTradeNo,
			"total_amount": amount,
			"subject":      req.Subject,
			"body":         omitEmpty(req.Body),
			"product_code": "QUICK_WAP_WAY",
			"time_expire":  omitEmpty(expire),
		}, req.NotifyURL, resp, dto.BodyLINK)
	case dto.MethodPC:
		err = payPage(client, "alipay.trade.page.pay", map[string]any{
			"out_trade_no": req.OutTradeNo,
			"total_amount": amount,
			"subject":      req.Subject,
			"body":         omitEmpty(req.Body),
			"product_code": "FAST_INSTANT_TRADE_PAY",
			"time_expire":  omitEmpty(expire),
		}, req.NotifyURL, resp, dto.BodyLINK)
	case dto.MethodAPP:
		err = paySdk(client, "alipay.trade.app.pay", map[string]any{
			"out_trade_no": req.OutTradeNo,
			"total_amount": amount,
			"subject":      req.Subject,
			"body":         omitEmpty(req.Body),
			"product_code": "QUICK_MSECURITY_PAY",
			"time_expire":  omitEmpty(expire),
		}, req.NotifyURL, resp)
	case dto.MethodQR:
		err = payQr(ctx, req, amount, expire, resp)
	case dto.MethodBARCODE:
		err = payBarcode(ctx, req, amount, expire, resp)
	case dto.MethodJSAPI:
		err = payJsapi(ctx, req, amount, expire, resp)
	default:
		return nil, alipay.NewSDKError("channel.error.alipayPayCallFailed", "unsupported method: "+string(req.Method))
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func payPage(client *openapi.Client, method string, biz map[string]any, notifyURL string, resp *dto.PayResp, bodyType dto.PayBodyType) error {
	cleanBiz(biz)
	url, err := client.PageExecuteGet(method, biz, notifyURL)
	if err != nil {
		return alipay.NewSDKError("channel.error.alipayPayCallFailed", err.Error())
	}
	resp.PayBody = url
	resp.PayBodyType = bodyType
	return nil
}

func paySdk(client *openapi.Client, method string, biz map[string]any, notifyURL string, resp *dto.PayResp) error {
	cleanBiz(biz)
	orderStr, err := client.SdkExecute(method, biz, notifyURL)
	if err != nil {
		return alipay.NewSDKError("channel.error.alipayPayCallFailed", err.Error())
	}
	resp.PayBody = orderStr
	resp.PayBodyType = dto.BodyOrderStr
	return nil
}

func payQr(ctx context.Context, req *dto.PayReq, amount, expire string, resp *dto.PayResp) error {
	biz := map[string]any{
		"out_trade_no": req.OutTradeNo,
		"total_amount": amount,
		"subject":      req.Subject,
		"body":         omitEmpty(req.Body),
		"time_expire":  omitEmpty(expire),
	}
	cleanBiz(biz)
	var out struct {
		gatewayBiz
		OutTradeNo string `json:"out_trade_no"`
		QRCode     string `json:"qr_code"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.precreate", biz, req.NotifyURL, &out); err != nil {
		return err
	}
	if !out.success() {
		return alipay.NewSDKError("channel.error.alipayPayCallFailed", out.errDetail())
	}
	resp.PayBody = out.QRCode
	resp.PayBodyType = dto.BodyQRCode
	return nil
}

func payBarcode(ctx context.Context, req *dto.PayReq, amount, expire string, resp *dto.PayResp) error {
	biz := map[string]any{
		"out_trade_no": req.OutTradeNo,
		"total_amount": amount,
		"subject":      req.Subject,
		"body":         omitEmpty(req.Body),
		"scene":        "bar_code",
		"auth_code":    req.AuthCode,
		"time_expire":  omitEmpty(expire),
	}
	cleanBiz(biz)
	var out struct {
		gatewayBiz
		TradeNo        string `json:"trade_no"`
		GmtPayment     string `json:"gmt_payment"`
		TotalAmount    string `json:"total_amount"`
		BuyerPayAmount string `json:"buyer_pay_amount"`
		ReceiptAmount  string `json:"receipt_amount"`
		BuyerUserID    string `json:"buyer_user_id"`
		BuyerOpenID    string `json:"buyer_open_id"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.pay", biz, req.NotifyURL, &out); err != nil {
		return err
	}
	if out.Code == codeSuccess {
		resp.TradeNo = out.TradeNo
		resp.Complete = true
		resp.FinishTime = alipay.ParseGatewayTime(out.GmtPayment)
		resp.TotalAmount = alipay.YuanToFenPtr(out.TotalAmount)
		resp.BuyerPayAmount = alipay.YuanToFenPtr(out.BuyerPayAmount)
		resp.ReceiptAmount = alipay.YuanToFenPtr(out.ReceiptAmount)
		resp.BuyerUserID = out.BuyerUserID
		resp.BuyerOpenID = out.BuyerOpenID
	}
	if out.Code != codeInProcess && !out.success() {
		return alipay.NewSDKError("channel.error.alipayPayCallFailed", out.errDetail())
	}
	return nil
}

func payJsapi(ctx context.Context, req *dto.PayReq, amount, expire string, resp *dto.PayResp) error {
	biz := map[string]any{
		"out_trade_no": req.OutTradeNo,
		"total_amount": amount,
		"subject":      req.Subject,
		"body":         omitEmpty(req.Body),
		"product_code": "JSAPI_PAY",
		"time_expire":  omitEmpty(expire),
	}
	if strings.HasPrefix(req.OpenID, buyerIDPrefix) {
		biz["buyer_id"] = req.OpenID
	} else {
		biz["op_buyer_open_id"] = req.OpenID
	}
	cleanBiz(biz)
	var out struct {
		gatewayBiz
		TradeNo    string `json:"trade_no"`
		OutTradeNo string `json:"out_trade_no"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.create", biz, req.NotifyURL, &out); err != nil {
		return err
	}
	if !out.success() {
		return alipay.NewSDKError("channel.error.alipayPayCallFailed", out.errDetail())
	}
	resp.TradeNo = out.TradeNo
	resp.PayBody = out.TradeNo
	resp.PayBodyType = dto.BodyIdentifier
	return nil
}

func omitEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func cleanBiz(biz map[string]any) {
	for k, v := range biz {
		if v == nil {
			delete(biz, k)
			continue
		}
		if s, ok := v.(string); ok && s == "" {
			delete(biz, k)
		}
	}
}

// Sync：查单
func Sync(ctx context.Context, req *dto.SyncReq) (*dto.SyncResp, error) {
	biz := map[string]any{"out_trade_no": req.OutTradeNo}
	if req.TradeNo != "" {
		biz["trade_no"] = req.TradeNo
	}
	var out struct {
		gatewayBiz
		TradeStatus    string `json:"trade_status"`
		TradeNo        string `json:"trade_no"`
		OutTradeNo     string `json:"out_trade_no"`
		SendPayDate    string `json:"send_pay_date"`
		BuyerUserID    string `json:"buyer_user_id"`
		BuyerOpenID    string `json:"buyer_open_id"`
		BuyerPayAmount string `json:"buyer_pay_amount"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.query", biz, "", &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayOrderQueryFailed")
	}
	resp := &dto.SyncResp{
		Code:        out.Code,
		SubCode:     out.SubCode,
		SubMsg:      out.SubMsg,
		TradeStatus: out.TradeStatus,
		TradeNo:     out.TradeNo,
		OutTradeNo:  out.OutTradeNo,
		SendPayDate: alipay.ParseGatewayTime(out.SendPayDate),
		BuyerUserID: out.BuyerUserID,
		BuyerOpenID: out.BuyerOpenID,
	}
	resp.BuyerPayAmount = alipay.YuanToFenPtr(out.BuyerPayAmount)
	return resp, nil
}

func wrapOr(err error, key string) error {
	if be, ok := err.(*alipay.BizError); ok {
		return alipay.NewSDKError(key, be.Detail)
	}
	return alipay.NewSDKError(key, err.Error())
}
