// Package openapi：银联商务开放平台协议层（原 sdk 包已更名为 openapi，与其它通道命名对齐）。
//
// 职责：HTTP 客户端、OPEN-BODY-SIG / OPEN-FORM-PARAM 签名、东八区时间、异步回调验签。
package openapi

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"daxpay.open/dax-pay-channel-one-go/internal/httpclient"
	"daxpay.open/dax-pay-channel-one-go/internal/ums"
)

const (
	// SandboxAPIURL：银联商务沙箱网关
	SandboxAPIURL = "https://test-api-open.chinaums.com"
	// ProductionAPIURL：银联商务生产网关
	ProductionAPIURL = "https://api-mop.chinaums.com"
)

// Client：银联商务 HTTP 客户端（对标 Boot UmsClient）
type Client struct {
	credential *ums.SdkCredential
	apiURL     string
	http       *http.Client
}

// NewClient：根据凭证构建客户端（sandbox 字段切换网关）
func NewClient(credential *ums.SdkCredential) *Client {
	base := ProductionAPIURL
	if credential != nil && credential.Sandbox {
		base = SandboxAPIURL
	}
	return &Client{
		credential: credential,
		apiURL:     strings.TrimRight(base, "/"),
		http:       httpclient.Default(),
	}
}

// QrPay：B 扫 C / 静态码获取二维码
func (c *Client) QrPay(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/bills/get-qrcode")
}

// QueryQrOrder：二维码订单查询
func (c *Client) QueryQrOrder(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/bills/query")
}

// RefundQr：二维码退款
func (c *Client) RefundQr(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/bills/refund")
}

// CloseQr：关闭二维码
func (c *Client) CloseQr(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/bills/close-qrcode")
}

// AlipayH5：支付宝 H5 跳转 URL（OPEN-FORM-PARAM）
func (c *Client) AlipayH5(param map[string]any) (string, error) {
	return c.buildH5URL(param, c.apiURL+"/v1/netpay/trade/h5-pay")
}

// WechatH5ToMini：微信 H5 转小程序跳转 URL
func (c *Client) WechatH5ToMini(param map[string]any) (string, error) {
	return c.buildH5URL(param, c.apiURL+"/v1/netpay/wxpay/h5-to-minipay")
}

// WechatH5：微信 H5 跳转 URL
func (c *Client) WechatH5(param map[string]any) (string, error) {
	return c.buildH5URL(param, c.apiURL+"/v1/netpay/wxpay/h5-pay")
}

// UnionH5：云闪付 H5 下单跳转 URL
func (c *Client) UnionH5(param map[string]any) (string, error) {
	return c.buildH5URL(param, c.apiURL+"/v1/netpay/uac/order")
}

// QueryH5Order：H5/线上订单查询
func (c *Client) QueryH5Order(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/query")
}

// QueryH5Refund：H5/线上退款查询
func (c *Client) QueryH5Refund(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/refund-query")
}

// RefundH5：H5/线上退款
func (c *Client) RefundH5(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/refund")
}

// CloseH5：H5/线上关单
func (c *Client) CloseH5(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/close")
}

// tradePost：JSON POST + OPEN-BODY-SIG；errCode!=SUCCESS 时包装为 BizError
func (c *Client) tradePost(param map[string]any, rawURL string) (map[string]any, error) {
	bodyBytes, err := json.Marshal(param)
	if err != nil {
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", err.Error())
	}
	body := string(bodyBytes)
	auth := OpenBodySig(c.credential.UmsAppID, c.credential.AppKey, body)

	req, err := http.NewRequest(http.MethodPost, rawURL, bytes.NewBufferString(body))
	if err != nil {
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", err.Error())
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", err.Error())
	}
	defer resp.Body.Close()
	resBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", err.Error())
	}

	var result map[string]any
	if err := json.Unmarshal(resBytes, &result); err != nil {
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", "invalid json: "+string(resBytes))
	}
	if strVal(result, "errCode") != "SUCCESS" {
		errMsg := strVal(result, "errMsg")
		if errMsg == "" {
			errMsg = "未知错误"
		}
		return nil, ums.NewSDKError("channel.error.umsRequestFailed", errMsg)
	}
	return result, nil
}

// buildH5URL：组装带 OPEN-FORM-PARAM 签名的浏览器跳转链接
func (c *Client) buildH5URL(param map[string]any, rawURL string) (string, error) {
	timestamp := H5Timestamp()
	nonce := randomDigits(32)
	bodyBytes, err := json.Marshal(param)
	if err != nil {
		return "", ums.NewSDKError("channel.error.umsRequestFailed", err.Error())
	}
	reqBody := string(bodyBytes)
	sig := Signature(c.credential.UmsAppID, c.credential.AppKey, timestamp, nonce, reqBody)
	return BuildH5URL(rawURL, c.credential.UmsAppID, timestamp, nonce, reqBody, sig), nil
}

func strVal(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		// JSON 数字（Unmarshal 默认 float64）
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case json.Number:
		return t.String()
	default:
		return fmt.Sprintf("%v", t)
	}
}

func randomDigits(n int) string {
	const digits = "0123456789"
	b := make([]byte, n)
	var buf [1]byte
	for i := range b {
		_, _ = rand.Read(buf[:])
		b[i] = digits[int(buf[0])%len(digits)]
	}
	return string(b)
}
