package sdk

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// OpenBodySig：生成 OPEN-BODY-SIG Authorization 头
func OpenBodySig(appID, appKey, body string) string {
	timestamp := H5Timestamp()
	nonce := newUUIDHex()
	sig := signature(appID, appKey, timestamp, nonce, body)
	return fmt.Sprintf(
		`OPEN-BODY-SIG AppId="%s", Timestamp="%s", Nonce="%s", Signature="%s"`,
		appID, timestamp, nonce, sig,
	)
}

// Signature：OPEN-FORM-PARAM 用 Base64 签名（timestamp/nonce 由调用方传入）
func Signature(appID, appKey, timestamp, nonce, body string) string {
	return signature(appID, appKey, timestamp, nonce, body)
}

func signature(appID, appKey, timestamp, nonce, body string) string {
	sum := sha256.Sum256([]byte(body))
	bodyDigest := hex.EncodeToString(sum[:])
	signContent := appID + timestamp + nonce + bodyDigest
	mac := hmac.New(sha256.New, []byte(appKey))
	_, _ = mac.Write([]byte(signContent))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// BuildH5URL：拼 OPEN-FORM-PARAM 跳转链接
func BuildH5URL(baseURL, appID, timestamp, nonce, reqBody, sig string) string {
	return fmt.Sprintf(
		"%s?authorization=OPEN-FORM-PARAM&appId=%s&timestamp=%s&nonce=%s&content=%s&signature=%s",
		baseURL, appID, timestamp, nonce,
		url.QueryEscape(reqBody),
		url.QueryEscape(sig),
	)
}

// VerifyCallback：异步回调验签
func VerifyCallback(params map[string]string, secretKey string) bool {
	if params == nil {
		return false
	}
	sign := params["sign"]
	if strings.TrimSpace(sign) == "" {
		return false
	}
	data := buildSignString(params)
	signType := params["signType"]
	var calculated string
	if signType == "MD5" {
		sum := md5.Sum([]byte(data + secretKey))
		calculated = hex.EncodeToString(sum[:])
	} else {
		sum := sha256.Sum256([]byte(data + secretKey))
		calculated = hex.EncodeToString(sum[:])
	}
	return strings.EqualFold(sign, calculated)
}

func buildSignString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || strings.TrimSpace(v) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return strings.Join(parts, "&")
}

func newUUIDHex() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
