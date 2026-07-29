package openapi

import "daxpay.open/dax-pay-channel-one-go/internal/alipay"

const (
	GatewayProd    = "https://openapi.alipay.com/gateway.do"
	GatewaySandbox = "https://openapi-sandbox.dl.alipaydev.com/gateway.do"
)

// GatewayURL：按凭证选择网关
//
// 沙箱开关最高优先：sandbox=true 时强制走沙箱网关，忽略自定义 ServerURL，
// 防止沙箱被绕过到自定义网关。非沙箱时允许 ServerURL 覆盖生产网关。
func GatewayURL(cred *alipay.SdkCredential) string {
	if cred != nil && cred.IsSandbox() {
		return GatewaySandbox
	}
	if cred != nil && cred.ServerURL != "" {
		return cred.ServerURL
	}
	return GatewayProd
}
