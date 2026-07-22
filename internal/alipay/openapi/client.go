package openapi

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
)

// Client：自研支付宝 OpenAPI 客户端
type Client struct {
	cred       *alipay.SdkCredential
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	httpClient *http.Client
	gateway    string
}

// NewClient：根据凭证构建客户端
func NewClient(cred *alipay.SdkCredential) (*Client, error) {
	if cred == nil {
		return nil, fmt.Errorf("credential is nil")
	}
	priv, err := ParsePrivateKey(cred.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	var pub *rsa.PublicKey
	if cred.IsCert() {
		pub, err = ExtractRSAPublicKey(cred.AlipayCert)
		if err != nil {
			return nil, fmt.Errorf("parse alipay cert: %w", err)
		}
	} else {
		pub, err = ParsePublicKey(cred.AlipayPublicKey)
		if err != nil {
			return nil, fmt.Errorf("parse alipay public key: %w", err)
		}
	}
	return &Client{
		cred:       cred,
		privateKey: priv,
		publicKey:  pub,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		gateway:    GatewayURL(cred),
	}, nil
}

// PublicKey：验签用支付宝公钥
func (c *Client) PublicKey() *rsa.PublicKey {
	return c.publicKey
}

// Execute：网关调用（需实际 HTTP），返回业务 JSON 对象（已验签）
func (c *Client) Execute(ctx context.Context, method string, biz any, notifyURL string) (json.RawMessage, error) {
	params, err := c.buildParams(method, biz, notifyURL)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.gateway, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return c.parseAndVerify(method, body)
}

// PageExecuteGet：本地签名拼 GET URL（WAP/PC）
func (c *Client) PageExecuteGet(method string, biz any, notifyURL string) (string, error) {
	params, err := c.buildParams(method, biz, notifyURL)
	if err != nil {
		return "", err
	}
	return c.gateway + "?" + EncodeQuery(params), nil
}

// SdkExecute：本地签名拼 orderStr（APP）
func (c *Client) SdkExecute(method string, biz any, notifyURL string) (string, error) {
	params, err := c.buildParams(method, biz, notifyURL)
	if err != nil {
		return "", err
	}
	return EncodeQuery(params), nil
}

func (c *Client) buildParams(method string, biz any, notifyURL string) (map[string]string, error) {
	bizJSON, err := json.Marshal(biz)
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"app_id":     c.cred.AliAppId,
		"method":     method,
		"format":     "JSON",
		"charset":    "utf-8",
		"sign_type":  c.cred.SignTypeOrDefault(),
		"timestamp":  alipay.NowCST(),
		"version":    "1.0",
		"biz_content": string(bizJSON),
	}
	if notifyURL != "" {
		params["notify_url"] = notifyURL
	}
	if c.cred.AppAuthToken != "" {
		params["app_auth_token"] = c.cred.AppAuthToken
	}
	if c.cred.IsCert() {
		appSN, err := CertSN(c.cred.AppCert)
		if err != nil {
			return nil, fmt.Errorf("app_cert_sn: %w", err)
		}
		rootSN, err := RootCertSN(c.cred.AlipayRootCert)
		if err != nil {
			return nil, fmt.Errorf("alipay_root_cert_sn: %w", err)
		}
		params["app_cert_sn"] = appSN
		params["alipay_root_cert_sn"] = rootSN
	}
	content := SignContent(params)
	sign, err := SignRSA2(content, c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	params["sign"] = sign
	return params, nil
}

func (c *Client) parseAndVerify(method string, body []byte) (json.RawMessage, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("parse response json: %w; body=%s", err, truncate(string(body), 200))
	}
	respKey := responseKey(method)
	bizRaw, ok := root[respKey]
	if !ok {
		// 错误时可能是 error_response
		if errRaw, ok2 := root["error_response"]; ok2 {
			return errRaw, nil
		}
		return nil, fmt.Errorf("missing %s in response: %s", respKey, truncate(string(body), 300))
	}
	var sign string
	if sRaw, ok := root["sign"]; ok {
		_ = json.Unmarshal(sRaw, &sign)
	}
	if sign != "" && c.publicKey != nil {
		// 验签原文为业务 JSON 原始字符串（支付宝约定）
		content := string(bizRaw)
		if err := VerifyRSA2(content, sign, c.publicKey); err != nil {
			return nil, fmt.Errorf("response verify failed: %w", err)
		}
	}
	return bizRaw, nil
}

func responseKey(method string) string {
	// alipay.trade.query → alipay_trade_query_response
	return strings.ReplaceAll(method, ".", "_") + "_response"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
