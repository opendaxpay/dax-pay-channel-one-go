package douyin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/openapi"
)

// JSAPISignMessage：JSAPI 二次签名原文（对标 Boot DouyinJsapiSigner）
func JSAPISignMessage(appID, timeStamp, nonceStr, prepayID string) string {
	return openapi.JSAPISignMessage(appID, timeStamp, nonceStr, prepayID)
}

// SignJSAPIPayInfo：计算 paySign（SHA256withRSA + Base64）
func SignJSAPIPayInfo(appID, timeStamp, nonceStr, prepayID, privateKeyRaw string) (string, error) {
	if appID == "" || timeStamp == "" || nonceStr == "" || prepayID == "" || privateKeyRaw == "" {
		return "", fmt.Errorf("抖音 JSAPI 签名参数存在空值")
	}
	sig, err := openapi.SignJSAPIPayInfo(appID, timeStamp, nonceStr, prepayID, privateKeyRaw)
	if err != nil {
		return "", fmt.Errorf("抖音 JSAPI paySign 签名失败: %w", err)
	}
	return sig, nil
}

// BuildJSAPIPayBody：组装前端 ttcjpay.dypay 所需 sdk_info JSON
func BuildJSAPIPayBody(appID, prepayID, privateKeyRaw string) (string, error) {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := openapi.NewNonce()
	paySign, err := SignJSAPIPayInfo(appID, timeStamp, nonce, prepayID, privateKeyRaw)
	if err != nil {
		return "", err
	}
	m := map[string]string{
		"appId":     appID,
		"timeStamp": timeStamp,
		"nonceStr":  nonce,
		"package":   "prepay_id=" + prepayID,
		"signType":  "DouyinPay-RSA",
		"paySign":   paySign,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
