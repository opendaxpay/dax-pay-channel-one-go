package openapi_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/wechat/openapi"
)

// TestEncryptOAEP：微信平台证书 RSA-OAEP(SHA-1) 加密 → 私钥解密还原
func TestEncryptOAEP(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := openapi.EncryptOAEP(&priv.PublicKey, "张三")
	if err != nil {
		t.Fatal(err)
	}
	cipher, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, priv, cipher, nil)
	if err != nil {
		t.Fatalf("decrypt OAEP failed: %v", err)
	}
	if string(plain) != "张三" {
		t.Fatalf("got %q, want 张三", string(plain))
	}
}

// TestEncryptOAEPNilKey：空公钥报错
func TestEncryptOAEPNilKey(t *testing.T) {
	if _, err := openapi.EncryptOAEP(nil, "x"); err == nil {
		t.Fatal("want error for nil key")
	}
}

// TestEncryptOAEPWithCert：从 PEM 证书提取公钥后加密(对齐实际使用路径)
func TestEncryptOAEPWithCert(t *testing.T) {
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
	encrypted, err := openapi.EncryptOAEP(pub, "收款人姓名")
	if err != nil {
		t.Fatal(err)
	}
	cipher, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, priv, cipher, nil)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if string(plain) != "收款人姓名" {
		t.Fatalf("got %q", string(plain))
	}
}
