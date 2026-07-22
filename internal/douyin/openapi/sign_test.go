package openapi_test

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/openapi"
)

func TestBuildSignMessage(t *testing.T) {
	msg := openapi.BuildSignMessage(
		"POST",
		"/v1/trade/transactions/native",
		"1710000000",
		"AbCdEfGhIjKlMnOpQrStUvWxYz012345",
		`{"appid":"tt123","mchid":"M1"}`,
	)
	want := "POST\n/v1/trade/transactions/native\n1710000000\nAbCdEfGhIjKlMnOpQrStUvWxYz012345\n{\"appid\":\"tt123\",\"mchid\":\"M1\"}\n"
	if msg != want {
		t.Fatalf("sign message mismatch:\n got %q\nwant %q", msg, want)
	}
}

func TestBuildVerifyMessage(t *testing.T) {
	msg := openapi.BuildVerifyMessage("1710000000", "nonce", `{"id":"1"}`)
	want := "1710000000\nnonce\n{\"id\":\"1\"}\n"
	if msg != want {
		t.Fatalf("verify message mismatch: %q", msg)
	}
}

func TestSignAndVerifyRSA(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	msg := openapi.BuildSignMessage("GET", "/v1/trade/transactions/out-trade-no/T1?mchid=M1", "1", "n", "")
	sig, err := openapi.SignRSA(priv, msg)
	if err != nil {
		t.Fatal(err)
	}
	if err := openapi.VerifyRSA(&priv.PublicKey, msg, sig); err != nil {
		t.Fatal(err)
	}
	auth := openapi.AuthorizationHeader("M1", "n", "1", "SERIAL", sig)
	if !strings.HasPrefix(auth, "DouyinPay-RSA ") {
		t.Fatalf("bad auth prefix: %s", auth)
	}
}

func TestJSAPISigner(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := rsaMarshalPKCS8(priv)
	if err != nil {
		t.Fatal(err)
	}
	pemKey := "-----BEGIN PRIVATE KEY-----\n" + chunk64(der) + "\n-----END PRIVATE KEY-----"

	appID := "tt_app"
	ts := "1710000000"
	nonce := "nonce32chars___________________"
	prepayID := "prepay_abc"
	msg := openapi.JSAPISignMessage(appID, ts, nonce, prepayID)
	lines := strings.Split(strings.TrimSuffix(msg, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("want 4 lines, got %d", len(lines))
	}
	sig, err := openapi.SignJSAPIPayInfo(appID, ts, nonce, prepayID, pemKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := openapi.VerifyRSA(&priv.PublicKey, msg, sig); err != nil {
		t.Fatal(err)
	}
	body, err := openapi.BuildJSAPIPayBody(appID, prepayID, pemKey)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"appId"`, `"timeStamp"`, `"nonceStr"`, `"package"`, `"signType"`, `"paySign"`, `"DouyinPay-RSA"`} {
		if !strings.Contains(body, key) {
			t.Fatalf("jsapi body missing %s: %s", key, body)
		}
	}
}

func TestAESGCMRoundTrip(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	nonce := "0123456789ab"
	plain := []byte(`{"out_trade_no":"T1"}`)
	ct, err := openapi.EncryptAESGCM(key, nonce, "transaction", plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := openapi.DecryptAESGCM(key, nonce, "transaction", ct)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Fatalf("plain mismatch: %s", got)
	}
}

func TestParsePrivateKeyBareBase64(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := rsaMarshalPKCS8(priv)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := openapi.ParsePrivateKey(encodeStd(der))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(priv.N) != 0 {
		t.Fatal("key mismatch")
	}
}
