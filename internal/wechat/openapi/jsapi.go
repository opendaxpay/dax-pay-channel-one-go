package openapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"
)

// SignJSAPIPayBody：JSAPI/MINI 调起参数 JSON
//
// 字段: appId, timeStamp, nonceStr, package, signType, paySign
func SignJSAPIPayBody(appID, prepayID string, signFn func(message string) (string, error)) (string, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce, err := randomNonce(16)
	if err != nil {
		return "", err
	}
	pkg := "prepay_id=" + prepayID
	msg := JSAPISignMessage(appID, ts, nonce, pkg)
	paySign, err := signFn(msg)
	if err != nil {
		return "", err
	}
	m := map[string]string{
		"appId":     appID,
		"timeStamp": ts,
		"nonceStr":  nonce,
		"package":   pkg,
		"signType":  "RSA",
		"paySign":   paySign,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// SignAppPayBody：APP 调起参数 JSON
//
// 字段: appid, partnerid, prepayid, package, noncestr, timestamp, sign
func SignAppPayBody(appID, partnerID, prepayID string, signFn func(message string) (string, error)) (string, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce, err := randomNonce(16)
	if err != nil {
		return "", err
	}
	msg := AppSignMessage(appID, ts, nonce, prepayID)
	sign, err := signFn(msg)
	if err != nil {
		return "", err
	}
	m := map[string]string{
		"appid":     appID,
		"partnerid": partnerID,
		"prepayid":  prepayID,
		"package":   "Sign=WXPay",
		"noncestr":  nonce,
		"timestamp": ts,
		"sign":      sign,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JSAPISignMessage：构造 JSAPI 二次签名原文
func JSAPISignMessage(appID, timestamp, nonceStr, packageValue string) string {
	return appID + "\n" + timestamp + "\n" + nonceStr + "\n" + packageValue + "\n"
}

// AppSignMessage：构造 APP 二次签名原文
func AppSignMessage(appID, timestamp, nonceStr, prepayID string) string {
	return appID + "\n" + timestamp + "\n" + nonceStr + "\n" + prepayID + "\n"
}

func randomNonce(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
