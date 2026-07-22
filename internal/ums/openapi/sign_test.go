package openapi_test

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/ums/openapi"
)

func TestSignatureDeterministic(t *testing.T) {
	appID, appKey := "app", "key"
	ts, nonce, body := "20260101120000", "abc", `{"a":1}`
	got := openapi.Signature(appID, appKey, ts, nonce, body)

	sum := sha256.Sum256([]byte(body))
	content := appID + ts + nonce + hex.EncodeToString(sum[:])
	mac := hmac.New(sha256.New, []byte(appKey))
	_, _ = mac.Write([]byte(content))
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if got != want {
		t.Fatalf("signature mismatch: got %s want %s", got, want)
	}
}

func TestBuildH5URL(t *testing.T) {
	u := openapi.BuildH5URL("https://example.com/pay", "app", "ts", "nonce", `{"x":1}`, "sig+")
	if !strings.Contains(u, "authorization=OPEN-FORM-PARAM") {
		t.Fatal(u)
	}
	if !strings.Contains(u, "content=") || !strings.Contains(u, "signature=") {
		t.Fatal(u)
	}
}

func TestVerifyCallbackMD5(t *testing.T) {
	secret := "secret"
	params := map[string]string{
		"billNo":   "B1",
		"signType": "MD5",
		"amount":   "100",
	}
	data := "amount=100&billNo=B1&signType=MD5"
	sum := md5.Sum([]byte(data + secret))
	params["sign"] = hex.EncodeToString(sum[:])
	if !openapi.VerifyCallback(params, secret) {
		t.Fatal("md5 verify failed")
	}
	params["sign"] = "bad"
	if openapi.VerifyCallback(params, secret) {
		t.Fatal("expected fail")
	}
}

func TestVerifyCallbackSHA256(t *testing.T) {
	secret := "secret"
	params := map[string]string{
		"merOrderId": "O1",
		"signType":   "SHA256",
	}
	data := "merOrderId=O1&signType=SHA256"
	sum := sha256.Sum256([]byte(data + secret))
	params["sign"] = hex.EncodeToString(sum[:])
	if !openapi.VerifyCallback(params, secret) {
		t.Fatal("sha256 verify failed")
	}
}
