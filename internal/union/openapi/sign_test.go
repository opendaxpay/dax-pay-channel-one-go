package openapi

import "testing"

func TestSignContent(t *testing.T) {
	params := map[string]string{
		"b":         "2",
		"a":         "1",
		"signature": "xxx",
		"empty":     "",
	}
	got := SignContent(params)
	// 字典序、排除 signature 与空值
	want := "a=1&b=2"
	if got != want {
		t.Errorf("SignContent() = %q, want %q", got, want)
	}
}

func TestParseFormResponse(t *testing.T) {
	body := "respCode=00&respMsg=ok&txnAmt=100"
	got := ParseFormResponse(body)
	if got["respCode"] != "00" {
		t.Errorf("respCode = %q, want 00", got["respCode"])
	}
	if got["txnAmt"] != "100" {
		t.Errorf("txnAmt = %q, want 100", got["txnAmt"])
	}
}
