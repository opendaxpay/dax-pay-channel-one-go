package union

// SdkCredential：云闪付(直连银联 ACP)通道凭证（对标 Boot UnionSdkCredential）
type SdkCredential struct {
	MerID             string `json:"merId"`
	SignType          string `json:"signType"`
	CertSign          bool   `json:"certSign"`
	KeyPrivateCert    string `json:"keyPrivateCert"`
	KeyPrivateCertPwd string `json:"keyPrivateCertPwd"`
	ACPMiddleCert     string `json:"acpMiddleCert"`
	ACPRootCert       string `json:"acpRootCert"`
	Sandbox           bool   `json:"sandbox"`
}
