package service

import (
	"strings"
	"time"
	"unicode/utf8"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin"
	"daxpay.open/dax-pay-channel-one-go/internal/douyin/openapi"
	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
)

const (
	currencyCNY      = "CNY"
	h5TypeWap        = "Wap"
	errOrderNotExist = "ORDER_NOT_EXIST"
	tradeStateClosed = "CLOSED"
)

func newClient(cred *douyin.SdkCredential) (*openapi.Client, error) {
	if cred == nil {
		return nil, douyin.NewSDKError("channel.error.douyinPayFailed", "credential is required")
	}
	c, err := openapi.NewClient(openapi.Config{
		MchID:      cred.MchID,
		SerialNo:   cred.MerchantSerialNumber,
		PrivateKey: cred.MerchantPrivateKey,
		EncryptKey: cred.EncryptKey,
	})
	if err != nil {
		return nil, douyin.NewSDKError("channel.error.douyinPayFailed", err.Error())
	}
	return c, nil
}

func wrapAPIErr(messageKey string, err error) error {
	if err == nil {
		return nil
	}
	if be, ok := err.(*douyin.BizError); ok {
		return be
	}
	if apiErr, ok := err.(*openapi.APIError); ok {
		return douyin.NewSDKError(messageKey, apiErr.Error())
	}
	return douyin.NewSDKError(messageKey, err.Error())
}

func containsOrderNotExist(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errOrderNotExist)
}

func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max])
}

// formatExpire：抖音要求 RFC3339，强制东八区
func formatExpire(t jsonx.OffsetDateTime) string {
	cst := t.Time().In(time.FixedZone("CST", 8*3600))
	return cst.Format("2006-01-02T15:04:05") + "+08:00"
}

func ptrInt64(v int64) *jsonx.Int64String {
	x := jsonx.Int64String(v)
	return &x
}

func amountTotal(total int64) map[string]any {
	return map[string]any{
		"total":    total,
		"currency": currencyCNY,
	}
}

func refundAmount(refund, total int64) map[string]any {
	return map[string]any{
		"refund":   refund,
		"total":    total,
		"currency": currencyCNY,
	}
}

func buildPayBase(reqCred *douyin.SdkCredential, outTradeNo, description, notifyURL, clientIP string, amount int64, expired *jsonx.OffsetDateTime) map[string]any {
	body := map[string]any{
		"appid":        reqCred.DouyinAppID,
		"mchid":        reqCred.MchID,
		"description":  truncateRunes(description, 127),
		"out_trade_no": outTradeNo,
		"amount":       amountTotal(amount),
	}
	if notifyURL != "" {
		body["notify_url"] = notifyURL
	}
	if expired != nil {
		body["time_expire"] = formatExpire(*expired)
	}
	if clientIP != "" {
		body["scene_info"] = map[string]any{"payer_client_ip": clientIP}
	}
	return body
}
