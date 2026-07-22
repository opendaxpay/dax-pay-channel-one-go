package openapi_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/wechat/openapi"
)

func TestAuthorizationMessageAndSign(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	msg := openapi.BuildAuthorizationMessage(
		"POST",
		"/v3/pay/transactions/jsapi",
		"1554208460",
		"593BEC0C930BF1AFEB40B4A08C8FB242",
		`{"amount":{"total":100}}`,
	)
	wantPrefix := "POST\n/v3/pay/transactions/jsapi\n1554208460\n"
	if !strings.HasPrefix(msg, wantPrefix) {
		t.Fatalf("message prefix mismatch: %q", msg)
	}
	if !strings.HasSuffix(msg, "\n") {
		t.Fatal("message must end with newline")
	}
	sig, err := openapi.SignSHA256WithRSA(msg, priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := openapi.VerifySHA256WithRSA(msg, sig, &priv.PublicKey); err != nil {
		t.Fatal(err)
	}
	auth := openapi.FormatAuthorization("1900009191", "593BEC0C930BF1AFEB40B4A08C8FB242", sig, "1554208460", "SERIAL")
	if !strings.HasPrefix(auth, "WECHATPAY2-SHA256-RSA2048 ") {
		t.Fatalf("bad auth: %s", auth)
	}
	if !strings.Contains(auth, `mchid="1900009191"`) {
		t.Fatalf("missing mchid: %s", auth)
	}
}

func TestJSAPISecondarySign(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	appID := "wx123"
	ts := "1414561699"
	nonce := "5K8264ILTKCH16CQ2502SI8ZNMTM67VS"
	pkg := "prepay_id=wx201410272009395522657a690389285100"
	msg := openapi.JSAPISignMessage(appID, ts, nonce, pkg)
	lines := strings.Split(strings.TrimSuffix(msg, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("want 4 lines, got %d: %q", len(lines), msg)
	}
	sig, err := openapi.SignSHA256WithRSA(msg, priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := openapi.VerifySHA256WithRSA(msg, sig, &priv.PublicKey); err != nil {
		t.Fatal(err)
	}

	body, err := openapi.SignJSAPIPayBody(appID, "wx201410272009395522657a690389285100", func(m string) (string, error) {
		return openapi.SignSHA256WithRSA(m, priv)
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"appId"`, `"timeStamp"`, `"nonceStr"`, `"package"`, `"signType"`, `"paySign"`} {
		if !strings.Contains(body, key) {
			t.Fatalf("jsapi body missing %s: %s", key, body)
		}
	}
}

func TestAEADRoundTrip(t *testing.T) {
	// apiKeyV3 必须 32 字节
	apiKey := "0123456789abcdef0123456789abcdef"
	nonce := "0123456789ab" // 12 bytes
	aad := "transaction"
	plain := []byte(`{"out_trade_no":"T1","trade_state":"SUCCESS"}`)

	ct, err := openapi.EncryptAEAD(apiKey, aad, nonce, plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := openapi.DecryptAEAD(apiKey, aad, nonce, ct)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Fatalf("plain mismatch: %s", got)
	}
}

func TestParsePrivateKeyPEM(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	parsed, err := openapi.ParsePrivateKey(string(pemBytes))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.N.Cmp(priv.N) != 0 {
		t.Fatal("key mismatch")
	}
}
