// Package openapi：银联全渠道支付平台(ACP)协议层。
//
// 职责：HTTP 客户端(form 提交)、RSA2 证书签名、东八区时间、异步回调验签。
package openapi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"sort"
	"strings"

	pkcs12 "software.sslmate.com/src/go-pkcs12"

	"daxpay.open/dax-pay-channel-one-go/internal/union"
)

// LoadSignCert：解析 PKCS12 私钥证书, 返回签名证书(对应 Java KeyStore.getCertificate)
func LoadSignCert(cred *union.SdkCredential) (*x509.Certificate, error) {
	pfx, err := base64.StdEncoding.DecodeString(cred.KeyPrivateCert)
	if err != nil {
		return nil, err
	}
	_, cert, _, err := pkcs12.DecodeChain(pfx, cred.KeyPrivateCertPwd)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

// LoadPrivateKey：解析 PKCS12 私钥, 返回 RSA 私钥(对应 Java KeyStore.getKey)
func LoadPrivateKey(cred *union.SdkCredential) (*rsa.PrivateKey, error) {
	pfx, err := base64.StdEncoding.DecodeString(cred.KeyPrivateCert)
	if err != nil {
		return nil, err
	}
	privateKey, _, _, err := pkcs12.DecodeChain(pfx, cred.KeyPrivateCertPwd)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}
	return rsaKey, nil
}

// GetSignCertID：获取签名证书序列号(银联 certId)
func GetSignCertID(cred *union.SdkCredential) (string, error) {
	cert, err := LoadSignCert(cred)
	if err != nil {
		return "", err
	}
	return cert.SerialNumber.String(), nil
}

// SignContent：银联签名原文(key 字典序, 跳过空值与 signature)
func SignContent(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "signature" || v == "" {
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

// Sign：RSA2 签名(SHA256WithRSA), 返回 Base64
func Sign(params map[string]string, cred *union.SdkCredential) (string, error) {
	content := SignContent(params)
	privateKey, err := LoadPrivateKey(cred)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(content))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// VerifyCallback：验证银联异步回调签名
//
// 银联回调附带 signPubKeyCert(银联签名证书 Base64 DER), 用其公钥校验 signature。
func VerifyCallback(params map[string]string, _ *union.SdkCredential) bool {
	sign := params["signature"]
	signPubKeyCert := params["signPubKeyCert"]
	if sign == "" || signPubKeyCert == "" {
		return false
	}
	certBytes, err := base64.StdEncoding.DecodeString(signPubKeyCert)
	if err != nil {
		return false
	}
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return false
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return false
	}
	content := SignContent(params)
	sig, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}
	h := sha256.Sum256([]byte(content))
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig) == nil
}
