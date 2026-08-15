package openapi_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/openapi"
)

// TestEncryptPKCS1v15：抖音平台证书 RSA PKCS#1 v1.5 加密 → 私钥解密还原
//
// 对齐抖音 SDK RsaCryptor(`RSA/ECB/PKCS1Padding`), 用于转账/分账敏感字段加密。
func TestEncryptPKCS1v15(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := openapi.EncryptPKCS1v15(&priv.PublicKey, "张三")
	if err != nil {
		t.Fatal(err)
	}
	cipher, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, priv, cipher)
	if err != nil {
		t.Fatalf("decrypt PKCS1v15 failed: %v", err)
	}
	if string(plain) != "张三" {
		t.Fatalf("got %q, want 张三", string(plain))
	}
}

// TestEncryptPKCS1v15NilKey：空公钥报错
func TestEncryptPKCS1v15NilKey(t *testing.T) {
	if _, err := openapi.EncryptPKCS1v15(nil, "x"); err == nil {
		t.Fatal("want error for nil key")
	}
}

// TestEncryptPKCS1v15WithCert：从 PEM 证书提取公钥后加密(对齐实际使用路径)
func TestEncryptPKCS1v15WithCert(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	pub, _, err := openapi.ParseCertificate(string(certPEM))
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := openapi.EncryptPKCS1v15(pub, "收款人手机号13800000000")
	if err != nil {
		t.Fatal(err)
	}
	cipher, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, priv, cipher)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if string(plain) != "收款人手机号13800000000" {
		t.Fatalf("got %q", string(plain))
	}
}
