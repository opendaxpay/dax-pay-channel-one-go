// Package version 暴露产品版本信息
package version

// 版本号, 与主应用(Java)/前端保持一致, 随发版同步
//
// 默认内嵌当前版本, 本地 go run / 普通编译可直接读到正确值;
// CI 编译时可用 ldflags 覆盖(同时注入 commit/build 时间做二进制追溯):
//
//	go build -ldflags "\
//	  -X daxpay.open/dax-pay-channel-one-go/internal/version.Version=4.0.0-beta4 \
//	  -X daxpay.open/dax-pay-channel-one-go/internal/version.GitCommit=$(git rev-parse --short HEAD) \
//	  -X daxpay.open/dax-pay-channel-one-go/internal/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
//	  ./cmd/server
var (
	// Version 产品版本号, 与 monorepo 其他端对齐
	Version = "4.0.0-beta4"
	// GitCommit 构建对应的 git commit (短 hash), 默认 unknown, CI 注入
	GitCommit = "unknown"
	// BuildTime 二进制构建时间 (UTC), 默认 unknown, CI 注入
	BuildTime = "unknown"
)
