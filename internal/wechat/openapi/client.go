package openapi

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/httpclient"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat"
)

const (
	// DefaultBaseURL：微信支付 V3 正式网关
	DefaultBaseURL = "https://api.mch.weixin.qq.com"
)

// APIError：微信 V3 错误响应
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Body    string `json:"-"`
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	return e.Message
}

// Client：自研微信 V3 HTTP 客户端（无 WxJava / 官方 SDK）
type Client struct {
	cred       *wechat.SdkCredential
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey // 支付公钥模式
	httpClient *http.Client
	baseURL    string
	certs      *CertStore
}

// NewClient：根据凭证构建客户端
func NewClient(cred *wechat.SdkCredential) (*Client, error) {
	if cred == nil || !cred.Valid() {
		return nil, fmt.Errorf("invalid wechat credential")
	}
	return newClientCore(cred)
}

// NewCallbackClient：回调专用构建（不要求 appId）
//
// 回调验签+解密不依赖 wxAppId，故放宽为 ValidForCallback 校验。
func NewCallbackClient(cred *wechat.SdkCredential) (*Client, error) {
	if cred == nil || !cred.ValidForCallback() {
		return nil, fmt.Errorf("invalid wechat credential")
	}
	return newClientCore(cred)
}

// newClientCore：凭证已校验后的客户端构造
func newClientCore(cred *wechat.SdkCredential) (*Client, error) {
	priv, err := ParsePrivateKey(cred.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	c := &Client{
		cred:       cred,
		privateKey: priv,
		httpClient: httpclient.Default(),
		baseURL:    DefaultBaseURL,
		certs:      NewCertStore(),
	}
	if cred.UsePublicKeyMode() {
		pub, err := ParsePublicKey(cred.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
		c.publicKey = pub
	}
	return c, nil
}

// Credential：当前凭证
func (c *Client) Credential() *wechat.SdkCredential {
	return c.cred
}

// PrivateKey：商户私钥
func (c *Client) PrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

// PlatformCert：获取微信支付平台证书公钥(转账敏感字段加密用)
//
// 缓存未命中时下载 /v3/certificates 并取最新一张; 微信平台证书一般仅一张,
// 轮换期多张时按下载列表最后一张为准。
func (c *Client) PlatformCert(ctx context.Context) (*rsa.PublicKey, error) {
	if pub := c.certs.Latest(); pub != nil {
		return pub, nil
	}
	if _, err := c.downloadAndCacheCert(ctx, ""); err != nil {
		return nil, err
	}
	if pub := c.certs.Latest(); pub != nil {
		return pub, nil
	}
	return nil, fmt.Errorf("no platform cert cached")
}

// SignMessage：用商户私钥签名
func (c *Client) SignMessage(message string) (string, error) {
	return SignSHA256WithRSA(message, c.privateKey)
}

// Do：发起已鉴权的 V3 请求，成功时返回响应 body（可能为空，如关单 204）
func (c *Client) Do(ctx context.Context, method, path string, body any) ([]byte, error) {
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

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce, err := randomNonce(16)
	if err != nil {
		return nil, err
	}

	msg := BuildAuthorizationMessage(method, path, timestamp, nonce, bodyStr)
	sig, err := SignSHA256WithRSA(msg, c.privateKey)
	if err != nil {
		return nil, err
	}
	auth := FormatAuthorization(c.cred.WxMchId, nonce, sig, timestamp, c.cred.CertSerialNo)

	url := c.baseURL + path
	var reader io.Reader
	if bodyBytes != nil {
		reader = strings.NewReader(bodyStr)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "daxpay-channel-one-go")
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// 支付公钥模式：告知微信加密所用公钥 ID
	if c.cred.UsePublicKeyMode() {
		req.Header.Set("Wechatpay-Serial", c.cred.PublicKeyId)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 2xx：验签（无 Wechatpay-Signature 时跳过，如部分 204）
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if err := c.verifyResponse(resp, respBody); err != nil {
			return nil, err
		}
		return respBody, nil
	}

	apiErr := &APIError{Status: resp.StatusCode, Body: string(respBody)}
	_ = json.Unmarshal(respBody, apiErr)
	if apiErr.Message == "" {
		apiErr.Message = string(respBody)
	}
	return nil, apiErr
}

// verifyResponse：校验应答签名
func (c *Client) verifyResponse(resp *http.Response, body []byte) error {
	sig := resp.Header.Get("Wechatpay-Signature")
	if sig == "" {
		return nil
	}
	ts := resp.Header.Get("Wechatpay-Timestamp")
	nonce := resp.Header.Get("Wechatpay-Nonce")
	serial := resp.Header.Get("Wechatpay-Serial")
	msg := BuildVerifyMessage(ts, nonce, string(body))
	pub, err := c.resolveVerifyKey(resp.Request.Context(), serial)
	if err != nil {
		return err
	}
	return VerifySHA256WithRSA(msg, sig, pub)
}

// resolveVerifyKey：按 serial 解析验签公钥（支付公钥 / 平台证书缓存 / 下载）
func (c *Client) resolveVerifyKey(ctx context.Context, serial string) (*rsa.PublicKey, error) {
	if c.cred.UsePublicKeyMode() {
		// publicKeyId 非空 → 支付公钥模式
		if serial == "" || serial == c.cred.PublicKeyId {
			return c.publicKey, nil
		}
		// serial 偶发为平台证书时仍尝试下载
	}
	if pub := c.certs.Get(serial); pub != nil {
		return pub, nil
	}
	// 下载平台证书（下载请求自身也需验签；首次可先跳过或用已有证书）
	return c.downloadAndCacheCert(ctx, serial)
}

// downloadAndCacheCert：GET /v3/certificates 并缓存
func (c *Client) downloadAndCacheCert(ctx context.Context, wantSerial string) (*rsa.PublicKey, error) {
	const path = "/v3/certificates"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce, err := randomNonce(16)
	if err != nil {
		return nil, err
	}
	msg := BuildAuthorizationMessage(http.MethodGet, path, timestamp, nonce, "")
	sig, err := SignSHA256WithRSA(msg, c.privateKey)
	if err != nil {
		return nil, err
	}
	auth := FormatAuthorization(c.cred.WxMchId, nonce, sig, timestamp, c.cred.CertSerialNo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "daxpay-channel-one-go")
	if c.cred.UsePublicKeyMode() {
		req.Header.Set("Wechatpay-Serial", c.cred.PublicKeyId)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download certs: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download certs status=%d body=%s", resp.StatusCode, truncate(string(body), 200))
	}

	// 证书下载响应可用支付公钥验签；平台证书模式下首次无缓存时跳过验签（仅信任 TLS）
	if c.cred.UsePublicKeyMode() && c.publicKey != nil {
		sig := resp.Header.Get("Wechatpay-Signature")
		ts := resp.Header.Get("Wechatpay-Timestamp")
		n := resp.Header.Get("Wechatpay-Nonce")
		if sig != "" {
			if err := VerifySHA256WithRSA(BuildVerifyMessage(ts, n, string(body)), sig, c.publicKey); err != nil {
				return nil, fmt.Errorf("cert download verify: %w", err)
			}
		}
	}

	pub, err := c.certs.LoadPlatformCerts(c.cred.ApiKeyV3, body, wantSerial)
	if err != nil {
		return nil, err
	}
	if pub == nil {
		return nil, fmt.Errorf("platform cert serial not found: %s", wantSerial)
	}
	return pub, nil
}

// VerifyNotify：校验回调签名（Wechatpay-* headers）
func (c *Client) VerifyNotify(timestamp, nonce, body, signature, serial string) error {
	msg := BuildVerifyMessage(timestamp, nonce, body)
	pub, err := c.resolveVerifyKey(context.Background(), serial)
	if err != nil {
		return err
	}
	return VerifySHA256WithRSA(msg, signature, pub)
}

// DecryptResource：用 apiKeyV3 解密回调 resource
func (c *Client) DecryptResource(associatedData, nonce, ciphertext string) ([]byte, error) {
	return DecryptAEAD(c.cred.ApiKeyV3, associatedData, nonce, ciphertext)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
