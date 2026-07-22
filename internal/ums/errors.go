package ums

import (
	"context"
	"fmt"

	"daxpay.open/dax-pay-channel-one-go/internal/i18n"
)

// BizError：通道业务错误（对标 ChannelServiceException，code=10003）
type BizError struct {
	Code       int
	MessageKey string
	Detail     string
}

func (e *BizError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.MessageKey, e.Detail)
	}
	return e.MessageKey
}

// NewSDKError：SDK/网关失败
func NewSDKError(messageKey, detail string) *BizError {
	return &BizError{Code: 10003, MessageKey: messageKey, Detail: detail}
}

// LocalizedMsg：渲染本地化消息（支持 {0} 占位）
func LocalizedMsg(ctx context.Context, e *BizError) string {
	if e == nil {
		return ""
	}
	if e.Detail != "" {
		msg := i18n.T(ctx, e.MessageKey, e.Detail)
		if msg != e.MessageKey {
			return msg
		}
		base := i18n.T(ctx, e.MessageKey)
		if base == e.MessageKey {
			return e.Detail
		}
		return base + ": " + e.Detail
	}
	return i18n.T(ctx, e.MessageKey)
}
