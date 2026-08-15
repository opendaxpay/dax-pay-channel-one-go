package openapi

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"daxpay.open/dax-pay-channel-one-go/internal/httpclient"
)

const (
	BaseURL  = "https://api.douyinpay.com"
	certPath = "/v1/merchant/certificates/getPlatformCertificates"
)

// Config：客户端配置（不依赖父包，避免 import cycle）
type Config struct {
	MchID      string
	SerialNo   string
	PrivateKey string
	EncryptKey string
}

// Client：自研抖音 OpenAPI HTTP 客户端
type Client struct {
	cfg        Config
	privateKey *rsa.PrivateKey
	httpClient *http.Client
	baseURL    string
	certs      *CertStore
}

// NewClient：根据配置构建客户端
func NewClient(cfg Config) (*Client, error) {
	if cfg.MchID == "" || cfg.SerialNo == "" || cfg.PrivateKey == "" || cfg.EncryptKey == "" {
		return nil, fmt.Errorf("incomplete credential")
	}
	pk, err := ParsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	return &Client{
		cfg:        cfg,
		privateKey: pk,
		httpClient: httpclient.Default(),
		baseURL:    BaseURL,
		certs:      NewCertStore(),
	}, nil
}

// APIError：抖音网关 HTTP/业务错误
type APIError struct {
	HTTPStatus int
	Code       string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	if e.Message != "" {
		return e.Message
	}
	return e.Body
}

// Do：签名并请求
func (c *Client) Do(ctx context.Context, method, pathWithQuery string, body any) ([]byte, error) {
	return c.do(ctx, method, pathWithQuery, body, true, nil)
}

// DoWithHeaders：签名并请求(附加自定义请求头, 如转账/分账的 Douyinpay-Serial)
//
// 签名原文(method/path/timestamp/nonce/body)不受附加头影响, 与 [Do] 同机制。
func (c *Client) DoWithHeaders(ctx context.Context, method, pathWithQuery string, body any, headers map[string]string) ([]byte, error) {
	return c.do(ctx, method, pathWithQuery, body, true, headers)
}

func (c *Client) do(ctx context.Context, method, pathWithQuery string, body any, verifyResp bool, headers map[string]string) ([]byte, error) {
	var bodyBytes []byte
	var bodyStr string
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyStr = string(bodyBytes)
	}

	timestamp := NowUnix()
	nonce := NewNonce()
	msg := BuildSignMessage(method, pathWithQuery, timestamp, nonce, bodyStr)
	sig, err := SignRSA(c.privateKey, msg)
	if err != nil {
		return nil, err
	}
	auth := AuthorizationHeader(c.cfg.MchID, nonce, timestamp, c.cfg.SerialNo, sig)

	var reader io.Reader
	if bodyBytes != nil {
		reader = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+pathWithQuery, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "daxpay-channel-one-go")
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if verifyResp {
			if err := c.verifyResponse(ctx, resp, raw); err != nil {
				return nil, err
			}
		}
		return raw, nil
	}

	apiErr := &APIError{HTTPStatus: resp.StatusCode, Body: string(raw)}
	var errBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(raw, &errBody)
	apiErr.Code = errBody.Code
	apiErr.Message = errBody.Message
	if apiErr.Message == "" {
		apiErr.Message = string(raw)
	}
	return nil, apiErr
}

func (c *Client) verifyResponse(ctx context.Context, resp *http.Response, body []byte) error {
	sig := resp.Header.Get("Douyinpay-Signature")
	if sig == "" {
		return nil
	}
	ts := resp.Header.Get("Douyinpay-Timestamp")
	nonce := resp.Header.Get("Douyinpay-Nonce")
	serial := resp.Header.Get("Douyinpay-Serial")
	msg := BuildVerifyMessage(ts, nonce, string(body))
	pub, err := c.resolveVerifyKey(ctx, serial)
	if err != nil {
		return err
	}
	return VerifyRSA(pub, msg, sig)
}

func (c *Client) resolveVerifyKey(ctx context.Context, serial string) (*rsa.PublicKey, error) {
	if pub := c.certs.Get(serial); pub != nil {
		return pub, nil
	}
	return c.downloadAndCacheCert(ctx, serial)
}

func (c *Client) downloadAndCacheCert(ctx context.Context, wantSerial string) (*rsa.PublicKey, error) {
	raw, err := c.do(ctx, http.MethodGet, certPath, nil, false, nil)
	if err != nil {
		return nil, fmt.Errorf("download platform certs: %w", err)
	}
	pub, err := c.certs.LoadPlatformCerts(c.cfg.EncryptKey, raw, wantSerial)
	if err != nil {
		return nil, err
	}
	if pub == nil {
		return nil, fmt.Errorf("platform cert serial not found: %s", wantSerial)
	}
	return pub, nil
}

// PlatformCert：获取平台证书公钥与序列号(转账/分账敏感字段加密)
//
// 缓存未命中时下载 getPlatformCertificates 并取最新一张;
// 序列号格式为十六进制大写, 与 SDK PemUtil.getSerialNumber 一致, 用于 Douyinpay-Serial 头。
func (c *Client) PlatformCert(ctx context.Context) (*rsa.PublicKey, string, error) {
	if pub := c.certs.Latest(); pub != nil {
		return pub, c.certs.LatestSerial(), nil
	}
	if _, err := c.downloadAndCacheCert(ctx, ""); err != nil {
		return nil, "", err
	}
	if pub := c.certs.Latest(); pub != nil {
		return pub, c.certs.LatestSerial(), nil
	}
	return nil, "", fmt.Errorf("no platform cert cached")
}

// VerifyNotify：校验回调签名
func (c *Client) VerifyNotify(ctx context.Context, timestamp, nonce, body, signature, serial string) error {
	msg := BuildVerifyMessage(timestamp, nonce, body)
	pub, err := c.resolveVerifyKey(ctx, serial)
	if err != nil {
		return err
	}
	return VerifyRSA(pub, msg, signature)
}

// DecryptResource：用 encryptKey 解密回调 resource
func (c *Client) DecryptResource(associatedData, nonce, ciphertext string) ([]byte, error) {
	return DecryptAESGCM(c.cfg.EncryptKey, nonce, associatedData, ciphertext)
}
