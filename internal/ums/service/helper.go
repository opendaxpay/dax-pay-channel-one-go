package service

import (
	"encoding/json"
	"fmt"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
	"daxpay.open/dax-pay-channel-one-go/internal/ums"
	"daxpay.open/dax-pay-channel-one-go/internal/ums/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/ums/openapi"
)

func newClient(cred *ums.SdkCredential) (*openapi.Client, error) {
	if cred == nil {
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", "credential is required")
	}
	return openapi.NewClient(cred), nil
}

func baseParam(cred *ums.SdkCredential) map[string]any {
	return map[string]any{
		"requestTimestamp": openapi.NowDateTime(),
		"mid":              cred.MerchantNo,
		"tid":              cred.TerminalNo,
	}
}

func mapStr(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case json.Number:
		return t.String()
	case bool:
		return fmt.Sprintf("%v", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func asMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	if s, ok := v.(string); ok && s != "" {
		var m map[string]any
		if json.Unmarshal([]byte(s), &m) == nil {
			return m
		}
	}
	return nil
}

func formatBillDate(t *jsonx.OffsetDateTime) string {
	if t == nil {
		return ""
	}
	tt := time.Time(*t)
	return openapi.FormatCstDate(&tt)
}

func isQR(method dto.PayMethod) bool {
	return method == dto.PayMethodQRCode
}
