package service

import (
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/openapi"
)

const (
	h5SceneType      = "Wap"
	currencyCNY      = "CNY"
	statusSuccess    = "SUCCESS"
	statusClosed     = "CLOSED"
	errOrderNotExist = "ORDER_NOT_EXIST"
	errOrderClosed   = "ORDER_CLOSED"
	errUserPaying    = "USERPAYING"
	errSystemError   = "SYSTEMERROR"
	errOrderPaid     = "ORDERPAID"
	errAuthCodeUsed  = "AUTH_CODE_USED"
)

func newClient(cred *wechat.SdkCredential) (*openapi.Client, error) {
	if cred == nil || !cred.Valid() {
		return nil, wechat.NewConfigError("channel.error.wechatInvalidConfig")
	}
	c, err := openapi.NewClient(cred)
	if err != nil {
		return nil, wechat.NewConfigError("channel.error.wechatInvalidConfig")
	}
	return c, nil
}

// newCallbackClient：回调专用客户端构建（不要求 appId）
//
// 回调验签+解密不依赖 wxAppId，故放宽校验。供 [ParsePayCallback] / [ParseRefundCallback] 使用。
func newCallbackClient(cred *wechat.SdkCredential) (*openapi.Client, error) {
	if cred == nil || !cred.ValidForCallback() {
		return nil, wechat.NewConfigError("channel.error.wechatInvalidConfig")
	}
	c, err := openapi.NewCallbackClient(cred)
	if err != nil {
		return nil, wechat.NewConfigError("channel.error.wechatInvalidConfig")
	}
	return c, nil
}

// formatExpire：微信要求 yyyy-MM-dd'T'HH:mm:ss+08:00（无小数秒）
func formatExpire(t jsonx.OffsetDateTime) string {
	cst := t.Time().In(time.FixedZone("CST", 8*3600))
	return cst.Format("2006-01-02T15:04:05") + "+08:00"
}

func parseRFC3339(s string) *jsonx.OffsetDateTime {
	if s == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return nil
		}
	}
	odt := jsonx.OffsetDateTime(parsed.UTC())
	return &odt
}

// parseTransferTime：微信转账时间(RFC3339 带毫秒与东八区偏移, 如 2026-07-06T11:15:48.123+08:00) → OffsetDateTime
func parseTransferTime(s string) *jsonx.OffsetDateTime {
	if s == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02T15:04:05.000-07:00", s)
	if err != nil {
		return nil
	}
	odt := jsonx.OffsetDateTime(parsed.UTC())
	return &odt
}

// parseAllocTime：微信分账时间(兼容 RFC3339 与 yyyy-MM-dd HH:mm:ss 东八区) → OffsetDateTime
func parseAllocTime(s string) *jsonx.OffsetDateTime {
	if s == "" {
		return nil
	}
	// 先试 RFC3339(含 +08:00 偏移)
	if parsed, err := time.Parse(time.RFC3339, s); err == nil {
		odt := jsonx.OffsetDateTime(parsed.UTC())
		return &odt
	}
	// 再试 yyyy-MM-dd HH:mm:ss 东八区字面量
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.FixedZone("CST", 8*3600)); err == nil {
		odt := jsonx.OffsetDateTime(parsed.UTC())
		return &odt
	}
	return nil
}

// isTransferTerminal：微信转账终态(SUCCESS/FAIL/CANCELLED)
func isTransferTerminal(state string) bool {
	return state == "SUCCESS" || state == "FAIL" || state == "CANCELLED"
}

// truncateRunes：按字符截断(UTF-8 安全, 对齐 Java StrUtil.sub)
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func ptrInt64(v int64) *jsonx.Int64String {
	x := jsonx.Int64String(v)
	return &x
}

func wrapAPIErr(messageKey string, err error) error {
	if err == nil {
		return nil
	}
	if be, ok := err.(*wechat.BizError); ok {
		return be
	}
	// 订单已支付/付款码已被使用: 实际可能是之前请求已成功(结果未知), 归入结果未知由主应用查单确认
	if isPayResultUnknown(err) {
		return wechat.NewResultUnknown("channel.error.wechatPayResultUnknown", err.Error())
	}
	if apiErr, ok := err.(*openapi.APIError); ok {
		return wechat.NewSDKError(messageKey, apiErr.Error())
	}
	return wechat.NewSDKError(messageKey, err.Error())
}

// isPayResultUnknown：订单已支付/付款码已被使用，结果未知
func isPayResultUnknown(err error) bool {
	apiErr, ok := err.(*openapi.APIError)
	if !ok {
		return false
	}
	return apiErr.Code == errOrderPaid || apiErr.Code == errAuthCodeUsed
}

func isCodepayPaying(err error) bool {
	apiErr, ok := err.(*openapi.APIError)
	if !ok {
		return false
	}
	return apiErr.Code == errUserPaying || apiErr.Code == errSystemError
}

func isCloseSuccessFallback(err error) bool {
	apiErr, ok := err.(*openapi.APIError)
	if !ok {
		return false
	}
	return apiErr.Code == errOrderNotExist || apiErr.Code == errOrderClosed
}

func amountMap(total int64) map[string]any {
	return map[string]any{
		"total":    total,
		"currency": currencyCNY,
	}
}
