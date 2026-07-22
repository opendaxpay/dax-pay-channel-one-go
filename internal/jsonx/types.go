package jsonx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Int64String：序列化为 JSON 字符串，反序列化同时接受字符串或数字（对标 Jackson Long→String）
type Int64String int64

func (v Int64String) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(v), 10) + `"`), nil
}

func (v *Int64String) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*v = 0
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		if s == "" {
			*v = 0
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("Int64String: %w", err)
		}
		*v = Int64String(n)
		return nil
	}
	n, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("Int64String: %w", err)
	}
	*v = Int64String(n)
	return nil
}

// OffsetDateTime：UTC ISO 8601（对标 Jackson OffsetDateTime → yyyy-MM-ddTHH:mm:ssZ）
type OffsetDateTime time.Time

const utcLayout = "2006-01-02T15:04:05Z"

func (t OffsetDateTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t).UTC()
	if tt.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + tt.Format(utcLayout) + `"`), nil
}

func (t *OffsetDateTime) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) || len(data) == 0 {
		*t = OffsetDateTime{}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		*t = OffsetDateTime{}
		return nil
	}
	if parsed, err := time.Parse(utcLayout, s); err == nil {
		*t = OffsetDateTime(parsed.UTC())
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("OffsetDateTime: %w", err)
	}
	*t = OffsetDateTime(parsed.UTC())
	return nil
}

func (t OffsetDateTime) Time() time.Time {
	return time.Time(t)
}

// Unmarshal：默认忽略未知字段（对标 FAIL_ON_UNKNOWN_PROPERTIES=false）
func Unmarshal(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}
