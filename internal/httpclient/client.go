// Package httpclient：通道 HTTP 客户端共享层
//
// 四通道（alipay/wechat/ums/douyin）的 openapi.Client 复用同一个 *http.Client，
// 启用连接池（keep-alive + MaxIdleConnsPerHost），避免每请求新建 client 导致的
// TCP 连接无法复用、TIME_WAIT 堆积、DNS 反复解析。
package httpclient

import (
	"net/http"
	"time"
)

// 默认共享 client（带调优过的 Transport）
//
// 连接池参数对标 Java Boot RestClientConfiguration：
//   - 连接超时 5s、握手 10s
//   - 每主机最大空闲连接 50，全局 200
//   - 空闲连接 90s 超时
//   - 总请求超时 30s（对标各通道原 *http.Client{Timeout: 30s}）
var defaultClient = newClient(30 * time.Second)

// Default 返回共享的 *http.Client
func Default() *http.Client {
	return defaultClient
}

func newClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   50,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
		},
		Timeout: timeout,
	}
}
