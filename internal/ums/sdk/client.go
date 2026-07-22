package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/ums"
)

const (
	SandboxAPIURL     = "https://test-api-open.chinaums.com"
	ProductionAPIURL  = "https://api-mop.chinaums.com"
)

// Client：银联商务 HTTP 客户端（对标 Boot UmsClient）
type Client struct {
	credential *ums.SdkCredential
	apiURL     string
	http       *http.Client
}

// NewClient：根据凭证构建客户端
func NewClient(credential *ums.SdkCredential) *Client {
	base := ProductionAPIURL
	if credential != nil && credential.Sandbox {
		base = SandboxAPIURL
	}
	return &Client{
		credential: credential,
		apiURL:     strings.TrimRight(base, "/"),
		http:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) QrPay(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/bills/get-qrcode")
}

func (c *Client) QueryQrOrder(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/bills/query")
}

func (c *Client) RefundQr(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/bills/refund")
}

func (c *Client) CloseQr(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/bills/close-qrcode")
}

func (c *Client) AlipayH5(param map[string]any) (string, error) {
	return c.buildH5URL(param, c.apiURL+"/v1/netpay/trade/h5-pay")
}

func (c *Client) WechatH5ToMini(param map[string]any) (string, error) {
	return c.buildH5URL(param, c.apiURL+"/v1/netpay/wxpay/h5-to-minipay")
}

func (c *Client) WechatH5(param map[string]any) (string, error) {
	return c.buildH5URL(param, c.apiURL+"/v1/netpay/wxpay/h5-pay")
}

func (c *Client) UnionH5(param map[string]any) (string, error) {
	return c.buildH5URL(param, c.apiURL+"/v1/netpay/uac/order")
}

func (c *Client) QueryH5Order(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/query")
}

func (c *Client) QueryH5Refund(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/refund-query")
}

func (c *Client) RefundH5(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/refund")
}

func (c *Client) CloseH5(param map[string]any) (map[string]any, error) {
	return c.tradePost(param, c.apiURL+"/v1/netpay/close")
}

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
		// JSON 数字
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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = digits[r.Intn(10)]
	}
	return string(b)
}
