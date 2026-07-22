package openapi

// JSAPISignMessage：JSAPI 二次签名原文（末行含 \n）
func JSAPISignMessage(appID, timeStamp, nonceStr, prepayID string) string {
	return appID + "\n" + timeStamp + "\n" + nonceStr + "\n" + "prepay_id=" + prepayID + "\n"
}

// SignJSAPIPayInfo：JSAPI 调起 paySign（对标 DouyinJsapiSigner）
func SignJSAPIPayInfo(appID, timeStamp, nonceStr, prepayID, privateKeyPEM string) (string, error) {
	pk, err := ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}
	return SignRSA(pk, JSAPISignMessage(appID, timeStamp, nonceStr, prepayID))
}
