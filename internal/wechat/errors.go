package wechat

import "daxpay.open/dax-pay-channel-one-go/internal/channelerr"

// 本包 BizError / New*Error 等为 channelerr 的类型与函数别名。
// 保留是为了让 service 继续写 wechat.NewSDKError(...)，避免全仓改 import；
// 实现以 channelerr 为准，勿在此再复制结构体。
type BizError = channelerr.BizError

var (
	NewSDKError      = channelerr.NewSDKError
	NewValidateError = channelerr.NewValidateError
	NewConfigError   = channelerr.NewConfigError
	NewResultUnknown = channelerr.NewResultUnknown
	LocalizedMsg     = channelerr.LocalizedMsg
)
