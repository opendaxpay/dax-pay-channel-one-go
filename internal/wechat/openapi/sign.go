package openapi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// BuildAuthorizationMessage：构造 WECHATPAY2-SHA256-RSA2048 签名原文
//
//	HTTP_METHOD\n
//	URL（含 query）\n
//	TIMESTAMP\n
//	NONCE\n
//	BODY\n
func BuildAuthorizationMessage(method, urlPath, timestamp, nonce, body string) string {
	return method + "\n" + urlPath + "\n" + timestamp + "\n" + nonce + "\n" + body + "\n"
}

// BuildVerifyMessage：构造应答/回调用验签原文
//
//	TIMESTAMP\n
//	NONCE\n
//	BODY\n
func BuildVerifyMessage(timestamp, nonce, body string) string {
	return timestamp + "\n" + nonce + "\n" + body + "\n"
}

// FormatAuthorization：拼 Authorization 头
func FormatAuthorization(mchID, nonce, signature, timestamp, serialNo string) string {
	return fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		mchID, nonce, signature, timestamp, serialNo,
	)
}

// SignSHA256WithRSA：SHA256WithRSA 签名，返回 Base64
func SignSHA256WithRSA(message string, privateKey *rsa.PrivateKey) (string, error) {
	if privateKey == nil {
		return "", errors.New("nil private key")
	}
	sum := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// VerifySHA256WithRSA：SHA256WithRSA 验签
func VerifySHA256WithRSA(message, signatureBase64 string, publicKey *rsa.PublicKey) error {
	if publicKey == nil {
		return errors.New("nil public key")
	}
	sig, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	sum := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, sum[:], sig)
}

// ParsePrivateKey：兼容 PKCS8 / PKCS1 PEM 或裸 Base64
func ParsePrivateKey(raw string) (*rsa.PrivateKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty private key")
	}
	block, _ := pem.Decode([]byte(raw))
	var der []byte
	if block != nil {
		der = block.Bytes
	} else {
		b, err := base64.StdEncoding.DecodeString(stripKeyWhitespace(raw))
		if err != nil {
			return nil, fmt.Errorf("decode private key: %w", err)
		}
		der = b
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("private key is not RSA")
	}
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	return nil, errors.New("unsupported private key format")
}

// ParsePublicKey：公钥 PEM 或裸 Base64
func ParsePublicKey(raw string) (*rsa.PublicKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty public key")
	}
	block, _ := pem.Decode([]byte(raw))
	var der []byte
	if block != nil {
		der = block.Bytes
	} else {
		wrapped := "-----BEGIN PUBLIC KEY-----\n" + chunkBase64(stripKeyWhitespace(raw)) + "\n-----END PUBLIC KEY-----"
		block, _ = pem.Decode([]byte(wrapped))
		if block == nil {
			b, err := base64.StdEncoding.DecodeString(stripKeyWhitespace(raw))
			if err != nil {
				return nil, fmt.Errorf("decode public key: %w", err)
			}
			der = b
		} else {
			der = block.Bytes
		}
	}
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}
	return rsaPub, nil
}

// ParseCertificate：解析 PEM 证书并返回公钥与序列号（十六进制大写）
func ParseCertificate(pemCert string) (*rsa.PublicKey, string, error) {
	block, _ := pem.Decode([]byte(pemCert))
	if block == nil {
		return nil, "", errors.New("invalid certificate pem")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, "", err
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, "", errors.New("certificate public key is not RSA")
	}
	return pub, strings.ToUpper(fmt.Sprintf("%X", cert.SerialNumber)), nil
}

func stripKeyWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}

func chunkBase64(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i += 64 {
		end := i + 64
		if end > len(s) {
			end = len(s)
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s[i:end])
	}
	return b.String()
}
