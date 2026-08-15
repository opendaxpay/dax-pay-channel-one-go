package service

import "testing"

// TestMapIdentityType：平台收款人类型 → 支付宝身份类型
func TestMapIdentityType(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"user_id", "ALIPAY_USER_ID"},
		{"open_id", "ALIPAY_OPEN_ID"},
		{"login_name", "ALIPAY_LOGON_ID"},
		{"未知类型默认user_id", "ALIPAY_USER_ID"},
		{"", "ALIPAY_USER_ID"},
	}
	for _, c := range cases {
		if got := mapIdentityType(c.in); got != c.want {
			t.Fatalf("mapIdentityType(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestTruncateRunes：按字符截断(UTF-8 安全, 对齐 Java StrUtil.sub)
func TestTruncateRunes(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"转账标题内容", 4, "转账标题"},
		{"hello", 3, "hel"},
		{"短", 5, "短"},
		{"", 5, ""},
		{"中文emoji🀄", 2, "中文"},
	}
	for _, c := range cases {
		if got := truncateRunes(c.in, c.max); got != c.want {
			t.Fatalf("truncateRunes(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}
