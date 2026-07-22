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
	"net/url"
	"sort"
	"strings"
)

// SignContent：支付宝签名原文（key 排序，跳过空值与 sign）
func SignContent(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || v == "" {
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
		// 可能是无头尾的 Base64
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

// ParsePublicKey：支付宝公钥（PEM 或裸 Base64）
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

// SignRSA2：SHA256WithRSA 签名，返回 Base64
func SignRSA2(content string, privateKey *rsa.PrivateKey) (string, error) {
	h := sha256.Sum256([]byte(content))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// VerifyRSA2：验签
func VerifyRSA2(content, sign string, publicKey *rsa.PublicKey) error {
	sig, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return fmt.Errorf("decode sign: %w", err)
	}
	h := sha256.Sum256([]byte(content))
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, h[:], sig)
}

// RsaCheckV1：异步回调验签（参数 map 含 sign，排除 sign/sign_type）
func RsaCheckV1(params map[string]string, publicKey *rsa.PublicKey) error {
	sign := params["sign"]
	if sign == "" {
		return errors.New("missing sign")
	}
	copied := make(map[string]string, len(params))
	for k, v := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		copied[k] = v
	}
	content := SignContent(copied)
	return VerifyRSA2(content, sign, publicKey)
}

// EncodeQuery：签名后参数拼查询串（对标 Java URLEncoder，空格为 %20 而非 +）
func EncodeQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+alipayQueryEscape(params[k]))
	}
	return strings.Join(parts, "&")
}

func alipayQueryEscape(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
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
