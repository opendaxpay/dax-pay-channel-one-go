package openapi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

// BuildSignMessage：请求签名串（末尾含换行）
//
//	METHOD\n
//	PATH_WITH_QUERY\n
//	timestamp\n
//	nonce\n
//	body\n
func BuildSignMessage(method, pathWithQuery, timestamp, nonce, body string) string {
	return method + "\n" + pathWithQuery + "\n" + timestamp + "\n" + nonce + "\n" + body + "\n"
}

// BuildVerifyMessage：应答/回调解签串
//
//	timestamp\n
//	nonce\n
//	body\n
func BuildVerifyMessage(timestamp, nonce, body string) string {
	return timestamp + "\n" + nonce + "\n" + body + "\n"
}

// SignRSA：SHA256withRSA + Base64
func SignRSA(privateKey *rsa.PrivateKey, message string) (string, error) {
	sum := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// VerifyRSA：验签
func VerifyRSA(publicKey *rsa.PublicKey, message, signatureBase64 string) error {
	sig, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, sum[:], sig)
}

// ParsePrivateKey：PKCS8/PKCS1 PEM 或裸 Base64（对标抖音 SDK PemUtil）
func ParsePrivateKey(raw string) (*rsa.PrivateKey, error) {
	normalized := strings.TrimSpace(raw)
	normalized = strings.ReplaceAll(normalized, "-----BEGIN PRIVATE KEY-----", "")
	normalized = strings.ReplaceAll(normalized, "-----END PRIVATE KEY-----", "")
	normalized = strings.ReplaceAll(normalized, "-----BEGIN RSA PRIVATE KEY-----", "")
	normalized = strings.ReplaceAll(normalized, "-----END RSA PRIVATE KEY-----", "")
	normalized = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, normalized)
	der, err := base64.StdEncoding.DecodeString(normalized)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		k, err2 := x509.ParsePKCS1PrivateKey(der)
		if err2 != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		return k, nil
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not RSA private key")
	}
	return rsaKey, nil
}

// ParseCertificate：解析 PEM 证书，返回公钥与序列号（十六进制大写）
func ParseCertificate(pemCert string) (*rsa.PublicKey, string, error) {
	block, _ := pem.Decode([]byte(pemCert))
	if block == nil {
		return nil, "", fmt.Errorf("invalid certificate pem")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, "", err
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, "", fmt.Errorf("certificate public key is not RSA")
	}
	return pub, strings.ToUpper(fmt.Sprintf("%X", cert.SerialNumber)), nil
}

// NewNonce：32 位字母数字随机串（对标 SDK SecureRandom）
func NewNonce() string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// NowUnix：秒级时间戳字符串
func NowUnix() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// AuthorizationHeader：组装 DouyinPay-RSA Authorization
func AuthorizationHeader(mchID, nonce, timestamp, serial, signature string) string {
	return fmt.Sprintf(
		`DouyinPay-RSA mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		mchID, nonce, timestamp, serial, signature,
	)
}
