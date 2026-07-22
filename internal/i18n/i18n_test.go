package i18n_test

import (
	"context"
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/i18n"
)

func TestLoadAndT(t *testing.T) {
	ms, err := i18n.Load(i18n.FS, "i18n")
	if err != nil {
		t.Fatal(err)
	}
	i18n.SetGlobal(ms)

	zh := i18n.WithLocale(context.Background(), "zh-CN")
	en := i18n.WithLocale(context.Background(), "en-US")

	if got := i18n.T(zh, "channel.error.systemError"); got != "系统内部错误" {
		t.Fatalf("zh systemError: %q", got)
	}
	if got := i18n.T(en, "channel.error.systemError"); got != "Internal system error" {
		t.Fatalf("en systemError: %q", got)
	}
	detail := i18n.T(zh, "channel.error.sdkCallFailedWithDetail", "x")
	if detail != "SDK 调用失败: x" {
		t.Fatalf("detail: %q", detail)
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	cases := map[string]string{
		"zh-Hans":                 "zh-CN",
		"en":                      "en-US",
		"ja-JP,en;q=0.8":          "ja-JP",
		"en-US;q=0.5,zh-CN;q=0.9": "zh-CN",
		"":                        "zh-CN",
	}
	for in, want := range cases {
		if got := i18n.ParseAcceptLanguage(in); got != want {
			t.Fatalf("ParseAcceptLanguage(%q)=%q want %q", in, got, want)
		}
	}
}
