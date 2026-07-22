package alipay

// SdkCredential：通道凭证（对标 Boot AlipaySdkCredential）
type SdkCredential struct {
	AliAppId       string `json:"aliAppId"`
	PrivateKey     string `json:"privateKey"`
	AlipayPublicKey string `json:"alipayPublicKey"`
	AppCert        string `json:"appCert"`
	AlipayCert     string `json:"alipayCert"`
	AlipayRootCert string `json:"alipayRootCert"`
	ServerURL      string `json:"serverUrl"`
	SignType       string `json:"signType"`
	AuthType       string `json:"authType"`
	Sandbox        *bool  `json:"sandbox"`
	AppAuthToken   string `json:"appAuthToken"`
}

// IsCert：authType=cert 为证书模式，其余（含空）为公钥模式
func (c *SdkCredential) IsCert() bool {
	if c == nil {
		return false
	}
	return c.AuthType == "cert"
}

// SignTypeOrDefault：默认 RSA2
func (c *SdkCredential) SignTypeOrDefault() string {
	if c == nil || c.SignType == "" {
		return "RSA2"
	}
	return c.SignType
}

// IsSandbox：是否沙箱
func (c *SdkCredential) IsSandbox() bool {
	return c != nil && c.Sandbox != nil && *c.Sandbox
}
