package openapi

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
)

// EncryptPKCS1v15：用平台证书公钥 RSA PKCS#1 v1.5 加密敏感字段, 返回 Base64
//
// 对齐抖音 SDK RsaCryptor(`RSA/ECB/PKCS1Padding`), 用于转账 user_name/phone_number
// 与分账接收方 name 的加密上送。
func EncryptPKCS1v15(publicKey *rsa.PublicKey, plain string) (string, error) {
	if publicKey == nil {
		return "", errors.New("nil public key")
	}
	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(plain))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipher), nil
}
