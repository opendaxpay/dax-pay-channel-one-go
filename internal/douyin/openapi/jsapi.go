package openapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// JSAPI 二次签名与前端调起参数组装。
//
// 放在 openapi（对标 wechat/openapi/jsapi.go），避免父包 douyin 再依赖本包造成
// import cycle（父包 credential ↔ openapi Client）。

// JSAPISignMessage：JSAPI 二次签名原文（末行含 \n）。
// 格式：appId\ntimeStamp\nnonceStr\nprepay_id={prepayID}\n
func JSAPISignMessage(appID, timeStamp, nonceStr, prepayID string) string {
	return appID + "\n" + timeStamp + "\n" + nonceStr + "\n" + "prepay_id=" + prepayID + "\n"
}

// SignJSAPIPayInfo：用商户私钥对 JSAPI 原文做 RSA 签名，得到 paySign（对标 DouyinJsapiSigner）
func SignJSAPIPayInfo(appID, timeStamp, nonceStr, prepayID, privateKeyPEM string) (string, error) {
	if appID == "" || timeStamp == "" || nonceStr == "" || prepayID == "" || privateKeyPEM == "" {
		return "", fmt.Errorf("抖音 JSAPI 签名参数存在空值")
	}
	pk, err := ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("抖音 JSAPI paySign 签名失败: %w", err)
	}
	sig, err := SignRSA(pk, JSAPISignMessage(appID, timeStamp, nonceStr, prepayID))
	if err != nil {
		return "", fmt.Errorf("抖音 JSAPI paySign 签名失败: %w", err)
	}
	return sig, nil
}

// BuildJSAPIPayBody：组装前端 ttcjpay.dypay 所需 sdk_info JSON。
//
// 字段：appId / timeStamp / nonceStr / package(prepay_id=...) /
// signType=DouyinPay-RSA / paySign
func BuildJSAPIPayBody(appID, prepayID, privateKeyRaw string) (string, error) {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := NewNonce()
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
