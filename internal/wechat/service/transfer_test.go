package service

import (
	"testing"
	"time"
)

// TestParseTransferTime：微信转账时间(RFC3339 带毫秒与东八区偏移)解析
func TestParseTransferTime(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		want  string // UTC 表示, 空串表示期望 nil
		isNil bool
	}{
		{"带毫秒东八区", "2026-07-06T11:15:48.123+08:00", "2026-07-06T03:15:48.123Z", false},
		{"东八区整点", "2026-07-06T00:00:00.000+08:00", "2026-07-05T16:00:00Z", false},
		{"空串", "", "", true},
		{"非法格式", "2026/07/06 11:15:48", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := parseTransferTime(c.in)
			if c.isNil {
				if got != nil {
					t.Fatalf("want nil, got %v", got.Time())
				}
				return
			}
			if got == nil {
				t.Fatal("want parsed time, got nil")
			}
			want, err := time.Parse(time.RFC3339Nano, c.want)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Time().Equal(want) {
				t.Fatalf("got %v, want %v", got.Time().UTC(), want)
			}
		})
	}
}

// TestParseAllocTime：微信分账时间(兼容 RFC3339 与 yyyy-MM-dd HH:mm:ss 东八区)解析
func TestParseAllocTime(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		want  string // UTC 表示
		isNil bool
	}{
		{"RFC3339 带偏移", "2026-07-06T11:15:48+08:00", "2026-07-06T03:15:48Z", false},
		{"东八区无时区字面量", "2026-07-06 11:15:48", "2026-07-06T03:15:48Z", false},
		{"空串", "", "", true},
		{"非法格式", "boom", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := parseAllocTime(c.in)
			if c.isNil {
				if got != nil {
					t.Fatalf("want nil, got %v", got.Time())
				}
				return
			}
			if got == nil {
				t.Fatal("want parsed time, got nil")
			}
			want, err := time.Parse(time.RFC3339, c.want)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Time().Equal(want) {
				t.Fatalf("got %v, want %v", got.Time().UTC(), want)
			}
		})
	}
}

// TestIsTransferTerminal：微信转账终态判断(SUCCESS/FAIL/CANCELLED)
func TestIsTransferTerminal(t *testing.T) {
	cases := []struct {
		state string
		want  bool
	}{
		{"SUCCESS", true},
		{"FAIL", true},
		{"CANCELLED", true},
		{"ACCEPTED", false},
		{"PROCESSING", false},
		{"WAIT_USER_CONFIRM", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isTransferTerminal(c.state); got != c.want {
			t.Fatalf("isTransferTerminal(%q) = %v, want %v", c.state, got, c.want)
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
		{"hello", 3, "hel"},
		{"转账备注内容", 4, "转账备注"},
		{"中文emoji🀄", 3, "中文e"}, // 按码点截断, 不会截断半个 UTF-8 序列
		{"短文本", 10, "短文本"},
		{"", 5, ""},
	}
	for _, c := range cases {
		if got := truncateRunes(c.in, c.max); got != c.want {
			t.Fatalf("truncateRunes(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}
