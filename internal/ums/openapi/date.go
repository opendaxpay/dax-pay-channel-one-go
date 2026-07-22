package openapi

import (
	"time"
)

// cst：东八区固定偏移（银联商务时间字段无时区字面量）
var cst = time.FixedZone("CST", 8*3600)

// NowDateTime：yyyy-MM-dd HH:mm:ss（东八区）
func NowDateTime() string {
	return time.Now().In(cst).Format("2006-01-02 15:04:05")
}

// TodayDate：yyyy-MM-dd（东八区）
func TodayDate() string {
	return time.Now().In(cst).Format("2006-01-02")
}

// H5Timestamp：yyyyMMddHHmmss（东八区，签名用）
func H5Timestamp() string {
	return time.Now().In(cst).Format("20060102150405")
}

// FormatCstDate：UTC/Offset 时间 → 东八区 yyyy-MM-dd；nil/零值返回空
func FormatCstDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.In(cst).Format("2006-01-02")
}
