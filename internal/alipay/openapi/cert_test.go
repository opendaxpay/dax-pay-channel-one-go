package openapi_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay/openapi"
)

func TestRootCertSNSkipsUnparseable(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	rsaCertPEM := mustSelfSignedPEM(t, rsaKey, x509.SHA256WithRSA)

	// 人为塞入无法被 ParseCertificate 成功使用的垃圾 CERTIFICATE 块 + 合法 RSA 证书
	// （真实环境是 elliptic curve；此处用损坏 DER 模拟 skipInvalid）
	junk := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte{0x30, 0x03, 0x01, 0x02, 0x03},
	})
	combined := string(junk) + "\n" + rsaCertPEM

	sn, err := openapi.RootCertSN(combined)
	if err != nil {
		t.Fatal(err)
	}
	if sn == "" || strings.Contains(sn, "_") {
		// 仅一本 RSA 证书，SN 无下划线
		if sn == "" {
			t.Fatal("empty sn")
		}
	}

	// 纯 RSA 根证书也能算
	sn2, err := openapi.RootCertSN(rsaCertPEM)
	if err != nil {
		t.Fatal(err)
	}
	if sn2 != sn {
		t.Fatalf("sn mismatch %s vs %s", sn, sn2)
	}
}

func TestRootCertSNWithECAndRSA(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rsaPEM := mustSelfSignedPEM(t, rsaKey, x509.SHA256WithRSA)
	ecPEM := mustSelfSignedPEM(t, ecKey, x509.ECDSAWithSHA256)
	combined := ecPEM + "\n" + rsaPEM

	sn, err := openapi.RootCertSN(combined)
	if err != nil {
		t.Fatalf("RootCertSN should skip EC and keep RSA: %v", err)
	}
	onlyRSA, err := openapi.RootCertSN(rsaPEM)
	if err != nil {
		t.Fatal(err)
	}
	if sn != onlyRSA {
		t.Fatalf("got %s want %s", sn, onlyRSA)
	}
}

func mustSelfSignedPEM(t *testing.T, key any, sigAlg x509.SignatureAlgorithm) string {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:         true,
	}
	var pub any
	var priv any
	switch k := key.(type) {
	case *rsa.PrivateKey:
		pub, priv = &k.PublicKey, k
	case *ecdsa.PrivateKey:
		pub, priv = &k.PublicKey, k
	default:
		t.Fatalf("unsupported key")
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		t.Fatal(err)
	}
	_ = sigAlg
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}
