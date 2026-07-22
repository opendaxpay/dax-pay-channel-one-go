package openapi_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay/openapi"
)

func TestSignAndVerifyRSA2(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

	parsedPriv, err := openapi.ParsePrivateKey(string(privPEM))
	if err != nil {
		t.Fatal(err)
	}
	parsedPub, err := openapi.ParsePublicKey(string(pubPEM))
	if err != nil {
		t.Fatal(err)
	}

	params := map[string]string{
		"app_id":      "2021000000",
		"method":      "alipay.trade.query",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"biz_content": `{"out_trade_no":"T1"}`,
		"sign":        "should-skip",
	}
	content := openapi.SignContent(params)
	if content == "" || strings.Contains(content, "sign=") {
		t.Fatalf("bad content: %s", content)
	}
	sig, err := openapi.SignRSA2(content, parsedPriv)
	if err != nil {
		t.Fatal(err)
	}
	if err := openapi.VerifyRSA2(content, sig, parsedPub); err != nil {
		t.Fatal(err)
	}

	// 回调验签：签名原文排除 sign / sign_type（对标 rsaCheckV1）
	cb := map[string]string{
		"out_trade_no": "T1",
		"trade_status": "TRADE_SUCCESS",
		"sign_type":    "RSA2",
	}
	cbForSign := map[string]string{
		"out_trade_no": "T1",
		"trade_status": "TRADE_SUCCESS",
	}
	cbContent := openapi.SignContent(cbForSign)
	cbSig, err := openapi.SignRSA2(cbContent, parsedPriv)
	if err != nil {
		t.Fatal(err)
	}
	cb["sign"] = cbSig
	if err := openapi.RsaCheckV1(cb, parsedPub); err != nil {
		t.Fatal(err)
	}
}

func TestSignContentOrder(t *testing.T) {
	got := openapi.SignContent(map[string]string{
		"b": "2",
		"a": "1",
		"c": "",
	})
	if got != "a=1&b=2" {
		t.Fatalf("got %q", got)
	}
}

func TestEncodeQuerySpace(t *testing.T) {
	q := openapi.EncodeQuery(map[string]string{"subject": "hello world"})
	if q != "subject=hello%20world" {
		t.Fatalf("got %q", q)
	}
}
