// Package channelerr 提供四通道共享的业务错误类型。
//
// 对标 Boot ChannelServiceException：service 抛出 *BizError，
// handler 经 errors.As 识别后写入 HTTP 200 + DaxResult（code + 本地化 msg）。
package channelerr

import (
	"context"
	"fmt"

	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/i18n"
)

// BizError：通道业务错误（对标 Boot ChannelServiceException）
type BizError struct {
	// Code：业务错误码（如 10003 SDK 调用失败）
	Code int
	// MessageKey：i18n 消息 key（如 channel.error.sdkCallFailedWithDetail）
	MessageKey string
	// Detail：可选细节；有值时作为 {0} 占位或回退拼接到本地化文案
	Detail string
}

// Error：实现 error；日志/兜底用，HTTP 响应请走 LocalizedMsg
func (e *BizError) Error() string {
	if e == nil {
		return ""
	}
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.MessageKey, e.Detail)
	}
	return e.MessageKey
}

// NewSDKError：SDK/网关失败（code=10003）
func NewSDKError(messageKey, detail string) *BizError {
	return &BizError{Code: errcode.SDKCallFailed.Code, MessageKey: messageKey, Detail: detail}
}

// NewValidateError：参数校验失败（code=10007）
func NewValidateError(messageKey string) *BizError {
	return &BizError{Code: errcode.ValidateParams.Code, MessageKey: messageKey}
}

// NewConfigError：通道配置无效（code=10002）
func NewConfigError(messageKey string) *BizError {
	return &BizError{Code: errcode.InvalidConfig.Code, MessageKey: messageKey}
}

// NewResultUnknown：结果未知（code=10009）
//
// 付款码用户支付中/订单已支付/付款码已被使用等场景，实际可能已成功，结果未知，
// 由主应用保持处理中并查单确认最终状态，避免误判 FAIL 导致资金悬挂。
func NewResultUnknown(messageKey, detail string) *BizError {
	return &BizError{Code: errcode.ResultUnknown.Code, MessageKey: messageKey, Detail: detail}
}

// LocalizedMsg：按 Accept-Language 渲染本地化消息。
//
// 有 Detail 时优先 i18n.T(key, detail)（支持 {0}）；若 key 无占位/未命中翻译，
// 再回退为「翻译文案: detail」或纯 detail。
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
