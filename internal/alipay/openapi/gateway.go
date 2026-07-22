package openapi

import "daxpay.open/dax-pay-channel-one-go/internal/alipay"

const (
	GatewayProd    = "https://openapi.alipay.com/gateway.do"
	GatewaySandbox = "https://openapi-sandbox.dl.alipaydev.com/gateway.do"
)

// GatewayURL：按凭证选择网关
func GatewayURL(cred *alipay.SdkCredential) string {
	if cred != nil && cred.ServerURL != "" {
		return cred.ServerURL
	}
	if cred != nil && cred.IsSandbox() {
		return GatewaySandbox
	}
	return GatewayProd
}
