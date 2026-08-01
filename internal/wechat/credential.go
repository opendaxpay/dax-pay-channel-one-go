package wechat

import "strings"

// SdkCredential：微信 SDK 凭证（对标 Boot WechatSdkCredential）
//
// 鉴权模式:
//   - 支付公钥新模式(优先): publicKeyId 非空时启用
//   - 平台证书模式(兜底): publicKeyId 为空时下载并缓存平台证书
type SdkCredential struct {
	WxMchId      string `json:"wxMchId"`
	WxAppId      string `json:"wxAppId"`
	SubMchId     string `json:"subMchId"`
	SubAppId     string `json:"subAppId"`
	ApiKeyV3     string `json:"apiKeyV3"`
	PrivateKey   string `json:"privateKey"`
	PrivateCert  string `json:"privateCert"`
	CertSerialNo string `json:"certSerialNo"`
	PublicKey    string `json:"publicKey"`
	PublicKeyId  string `json:"publicKeyId"`
}

// UsePublicKeyMode：支付公钥新模式（publicKeyId 非空）
func (c *SdkCredential) UsePublicKeyMode() bool {
	return c != nil && strings.TrimSpace(c.PublicKeyId) != ""
}

// Valid：必填校验（wxMchId / wxAppId / apiKeyV3 / privateKey / certSerialNo）
func (c *SdkCredential) Valid() bool {
	if c == nil {
		return false
	}
	return strings.TrimSpace(c.WxMchId) != "" &&
		strings.TrimSpace(c.WxAppId) != "" &&
		strings.TrimSpace(c.ApiKeyV3) != "" &&
		strings.TrimSpace(c.PrivateKey) != "" &&
		strings.TrimSpace(c.CertSerialNo) != ""
}

// ValidForCallback：回调验签专用必填校验（不要求 wxAppId）
//
// 回调验签+解密仅需 apiKeyV3 与证书(平台证书模式还需 mchId 做证书下载鉴权), 不依赖 wxAppId。
func (c *SdkCredential) ValidForCallback() bool {
	if c == nil {
		return false
	}
	return strings.TrimSpace(c.WxMchId) != "" &&
		strings.TrimSpace(c.ApiKeyV3) != "" &&
		strings.TrimSpace(c.PrivateKey) != "" &&
		strings.TrimSpace(c.CertSerialNo) != ""
}
