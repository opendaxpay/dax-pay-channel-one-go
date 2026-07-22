package alipay_test

import (
	"testing"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
)

func TestFenYuanRoundTrip(t *testing.T) {
	if got := alipay.FenToYuan(1025); got != "10.25" {
		t.Fatalf("FenToYuan=%s", got)
	}
	fen, ok := alipay.YuanToFen("10.25")
	if !ok || fen != 1025 {
		t.Fatalf("YuanToFen=%d ok=%v", fen, ok)
	}
	fen, ok = alipay.YuanToFen("10.256")
	if !ok || fen != 1026 {
		t.Fatalf("HALF_UP YuanToFen=%d", fen)
	}
}

func TestFormatExpireCST(t *testing.T) {
	// 2026-07-06 03:15:48 UTC = 2026-07-06 11:15:48 CST
	tt := time.Date(2026, 7, 6, 3, 15, 48, 0, time.UTC)
	got := alipay.FormatExpire(jsonx.OffsetDateTime(tt))
	if got != "2026-07-06 11:15:48" {
		t.Fatalf("FormatExpire=%s", got)
	}
}

func TestParseCst(t *testing.T) {
	v := alipay.ParseCst("2026-07-06 11:15:48")
	if v == nil {
		t.Fatal("nil")
	}
	want := time.Date(2026, 7, 6, 3, 15, 48, 0, time.UTC)
	if !v.Time().Equal(want) {
		t.Fatalf("got %v want %v", v.Time(), want)
	}
}
