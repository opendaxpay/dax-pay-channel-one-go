package errcode

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/i18n"
)

// 通道服务错误码（对标 Boot ChannelErrorCode）
//
// code 为业务错误码，messageKey 对应 i18n 资源 channel/error.json 中的 key。
type ChannelError struct {
	Code       int
	MessageKey string
}

var (
	Success                 = ChannelError{0, "channel.error.success"}
	ChannelNotFound         = ChannelError{10001, "channel.error.channelNotFound"}
	InvalidConfig           = ChannelError{10002, "channel.error.invalidConfig"}
	SDKCallFailed           = ChannelError{10003, "channel.error.sdkCallFailed"}
	CallbackVerifyFailed    = ChannelError{10004, "channel.error.callbackVerifyFailed"}
	CacheConcurrentConflict = ChannelError{10005, "channel.error.cacheConcurrentConflict"}
	SystemError             = ChannelError{10006, "channel.error.systemError"}
	ValidateParams          = ChannelError{10007, "channel.error.validateParams"}
	ResponseVerifyFailed    = ChannelError{10008, "channel.error.responseVerifyFailed"}
	// 结果未知(用户支付中/付款码已使用/订单已支付等, 需主应用查单确认最终状态)
	ResultUnknown = ChannelError{10009, "channel.error.resultUnknown"}
)

// 获取当前 locale 下的本地化消息
func (e ChannelError) Message(ctx context.Context) string {
	return i18n.T(ctx, e.MessageKey)
}

// 带 {0} 占位符的本地化消息（如 channel.error.sdkCallFailedWithDetail）
func (e ChannelError) MessageWithDetail(ctx context.Context, detail string) string {
	key := e.MessageKey + "WithDetail"
	msg := i18n.T(ctx, key, detail)
	if msg == key {
		return e.Message(ctx) + ": " + detail
	}
	return msg
}
