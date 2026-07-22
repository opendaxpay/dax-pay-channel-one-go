package i18n

import "embed"

// 嵌入十语 channel/error.json（从 Boot common-i18n 拷贝）
//
//go:embed all:i18n
var FS embed.FS
