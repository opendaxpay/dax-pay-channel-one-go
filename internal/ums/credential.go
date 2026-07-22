package ums

// SdkCredential：银联商务通道凭证（对标 Boot UmsSdkCredential）
type SdkCredential struct {
	UmsAppID   string `json:"umsAppId"`
	AppKey     string `json:"appKey"`
	MerchantNo string `json:"merchantNo"`
	TerminalNo string `json:"terminalNo"`
	SecretKey  string `json:"secretKey"`
	Sandbox    bool   `json:"sandbox"`
}
