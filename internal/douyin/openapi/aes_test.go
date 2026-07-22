package openapi_test

import (
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/openapi"
)

func TestDecryptAESGCMKnownVector(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef" // 32
	nonce := "0123456789ab"                     // 12
	aad := "certificate"
	plain := []byte("hello-douyin-aead")

	if len(key) != 32 {
		t.Fatalf("key len %d", len(key))
	}
	if len(nonce) != 12 {
		t.Fatalf("nonce len %d", len(nonce))
	}

	ct, err := openapi.EncryptAESGCM(key, nonce, aad, plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := openapi.DecryptAESGCM(key, nonce, aad, ct)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Fatalf("got %q want %q", got, plain)
	}
}

func TestDecryptAESGCMBadKeyLen(t *testing.T) {
	_, err := openapi.DecryptAESGCM("short", "0123456789ab", "", "AAAA")
	if err == nil {
		t.Fatal("expected key length error")
	}
}
