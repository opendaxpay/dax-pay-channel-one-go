package openapi

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"errors"
)

// EncryptOAEP：用微信支付平台证书公钥 RSA-OAEP(SHA-1, MGF1) 加密敏感字段, 返回 Base64
//
// 对齐 WxJava @SpecEncrypt 的 `RSA/ECB/OAEPWithSHA-1AndMGF1Padding`(商家转账到零钱 user_name)。
func EncryptOAEP(publicKey *rsa.PublicKey, plain string) (string, error) {
	if publicKey == nil {
		return "", errors.New("nil public key")
	}
	cipher, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, publicKey, []byte(plain), nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipher), nil
}
