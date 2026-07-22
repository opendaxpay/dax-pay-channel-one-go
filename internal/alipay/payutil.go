package alipay

import (
	"math/big"
	"strings"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
)

var (
	fenPerYuan = big.NewRat(100, 1)
	cst        = time.FixedZone("CST", 8*3600)
)

const cstLayout = "2006-01-02 15:04:05"

// FenToYuan：分 → 元字符串（保留两位，HALF_UP）
func FenToYuan(fen int64) string {
	r := new(big.Rat).SetInt64(fen)
	r.Quo(r, fenPerYuan)
	return r.FloatString(2)
}

// YuanToFen：元字符串 → 分（HALF_UP）；空白返回 0,false
func YuanToFen(yuan string) (int64, bool) {
	yuan = strings.TrimSpace(yuan)
	if yuan == "" {
		return 0, false
	}
	r := new(big.Rat)
	if _, ok := r.SetString(yuan); !ok {
		return 0, false
	}
	r.Mul(r, fenPerYuan)
	// HALF_UP：加 0.5 再向零取整（正数）
	half := big.NewRat(1, 2)
	r.Add(r, half)
	f, _ := r.Float64()
	return int64(f), true
}

// YuanToFenPtr：空白返回 nil
func YuanToFenPtr(yuan string) *jsonx.Int64String {
	fen, ok := YuanToFen(yuan)
	if !ok {
		return nil
	}
	v := jsonx.Int64String(fen)
	return &v
}

// FormatExpire：UTC OffsetDateTime → 东八区 yyyy-MM-dd HH:mm:ss（支付宝 time_expire）
func FormatExpire(t jsonx.OffsetDateTime) string {
	tt := t.Time()
	if tt.IsZero() {
		return ""
	}
	return tt.In(cst).Format(cstLayout)
}

// ParseCst：东八区本地时间字面量 → OffsetDateTime
func ParseCst(text string) *jsonx.OffsetDateTime {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	tt, err := time.ParseInLocation(cstLayout, text, cst)
	if err != nil {
		return nil
	}
	v := jsonx.OffsetDateTime(tt.UTC())
	return &v
}

// ParseGatewayTime：网关返回的时间（可能是 CST 字面量或 RFC3339）→ OffsetDateTime
func ParseGatewayTime(text string) *jsonx.OffsetDateTime {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if v := ParseCst(text); v != nil {
		return v
	}
	if tt, err := time.Parse(time.RFC3339, text); err == nil {
		v := jsonx.OffsetDateTime(tt.UTC())
		return &v
	}
	return nil
}

// NowCST：当前东八区时间字符串（公共参数 timestamp）
func NowCST() string {
	return time.Now().In(cst).Format(cstLayout)
}
