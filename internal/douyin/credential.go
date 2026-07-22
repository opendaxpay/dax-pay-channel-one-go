package douyin

// SdkCredential：抖音通道凭证（对标 Boot DouyinSdkCredential）
type SdkCredential struct {
	DouyinAppID           string `json:"douyinAppId"`
	MchID                 string `json:"mchId"`
	MerchantSerialNumber  string `json:"merchantSerialNumber"`
	MerchantPrivateKey    string `json:"merchantPrivateKey"`
	EncryptKey            string `json:"encryptKey"`
}
