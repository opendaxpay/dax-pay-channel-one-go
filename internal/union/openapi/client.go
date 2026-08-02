package openapi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"daxpay.open/dax-pay-channel-one-go/internal/httpclient"
	"daxpay.open/dax-pay-channel-one-go/internal/union"
)

const (
	// SandboxHost：银联沙箱网关
	SandboxHost = "test.95516.com"
	// ProductionHost：银联生产网关
	ProductionHost = "95516.com"

	backTransURL  = "https://gateway.%s/gateway/api/backTransReq.do"
	frontTransURL = "https://gateway.%s/gateway/api/frontTransReq.do"
	queryTransURL = "https://gateway.%s/gateway/api/queryTrans.do"
)

// Client：云闪付(直连银联 ACP) HTTP 客户端(对标 Boot UnionClient)
type Client struct {
	credential *union.SdkCredential
	host       string
	http       *http.Client
}

// NewClient：根据凭证构建客户端(sandbox 字段切换网关)
func NewClient(credential *union.SdkCredential) *Client {
	host := ProductionHost
	if credential != nil && credential.Sandbox {
		host = SandboxHost
	}
	return &Client{
		credential: credential,
		host:       host,
		http:       httpclient.Default(),
	}
}

// ApplyQRCode：主扫支付(申请二维码, C 扫 B), 响应中取 qrNo
func (c *Client) ApplyQRCode(param map[string]string) (map[string]string, error) {
	param["txnType"] = "01"
	param["txnSubType"] = "07"
	param["bizType"] = "000000"
	return c.tradePost(param, fmt.Sprintf(backTransURL, c.host))
}

// Consume：被扫支付(付款码消费, B 扫 C), 需传入 qrNo(用户付款码)
func (c *Client) Consume(param map[string]string) (map[string]string, error) {
	param["txnType"] = "01"
	param["txnSubType"] = "06"
	param["bizType"] = "000000"
	return c.tradePost(param, fmt.Sprintf(backTransURL, c.host))
}

// QueryTrans：单笔交易查询(支付/退款状态通用)
func (c *Client) QueryTrans(param map[string]string) (map[string]string, error) {
	return c.tradePost(param, fmt.Sprintf(queryTransURL, c.host))
}

// Refund：退款(退货), 需传入 origQryId(原交易凭证)
func (c *Client) Refund(param map[string]string) (map[string]string, error) {
	param["txnType"] = "04"
	param["txnSubType"] = "00"
	param["bizType"] = "000000"
	return c.tradePost(param, fmt.Sprintf(backTransURL, c.host))
}

// CloseOrder：关闭订单(交易撤销/关闭), 需传入 origQryId
func (c *Client) CloseOrder(param map[string]string) (map[string]string, error) {
	param["txnType"] = "31"
	param["txnSubType"] = "00"
	param["bizType"] = "000000"
	return c.tradePost(param, fmt.Sprintf(backTransURL, c.host))
}

// BuildWapFormHTML：构建 H5/WAP 前台跳转自动提交表单
//
// 银联前台交易必须由浏览器 POST 到 frontTransReq.do(不支持 GET),
// 返回包含隐藏表单 + 自动 submit 脚本的 HTML, 主应用前端 document.write 跳转。
func (c *Client) BuildWapFormHTML(param map[string]string) (string, error) {
	param["txnType"] = "01"
	param["txnSubType"] = "01"
	param["bizType"] = "000201"
	if err := c.signParam(param); err != nil {
		return "", err
	}
	action := fmt.Sprintf(frontTransURL, c.host)
	return buildAutoSubmitForm(action, param), nil
}

// tradePost：通用后台 POST form 提交(自动 RSA2 签名 + 响应解析)
func (c *Client) tradePost(param map[string]string, rawURL string) (map[string]string, error) {
	if err := c.signParam(param); err != nil {
		return nil, union.NewSDKError("channel.error.unionSignFailed", err.Error())
	}
	form := url.Values{}
	for k, v := range param {
		form.Set(k, v)
	}
	req, err := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", err.Error())
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", err.Error())
	}
	defer resp.Body.Close()
	resBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", err.Error())
	}
	result := ParseFormResponse(string(resBytes))
	if len(result) == 0 {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", "空响应")
	}
	if _, ok := result["respCode"]; !ok {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", string(resBytes))
	}
	return result, nil
}

// signParam：对报文参数补全签名信息(signMethod/certId/signature)
//
// certId 必须在签名前放入(签名内容包含 certId), signature 最后放入(不参与签名)
func (c *Client) signParam(param map[string]string) error {
	param["signMethod"] = "01"
	certID, err := GetSignCertID(c.credential)
	if err != nil {
		return err
	}
	param["certId"] = certID
	sign, err := Sign(param, c.credential)
	if err != nil {
		return err
	}
	param["signature"] = sign
	return nil
}

// ParseFormResponse：解析银联 form 格式响应(key=value&key=value)
func ParseFormResponse(body string) map[string]string {
	result := make(map[string]string)
	for _, kv := range strings.Split(body, "&") {
		idx := strings.Index(kv, "=")
		if idx > 0 {
			key := kv[:idx]
			value, err := url.QueryUnescape(kv[idx+1:])
			if err != nil {
				value = kv[idx+1:]
			}
			result[key] = value
		}
	}
	return result
}

// buildAutoSubmitForm：生成自动提交的 HTML 表单(银联前台跳转)
func buildAutoSubmitForm(action string, param map[string]string) string {
	var b strings.Builder
	b.WriteString(`<form action="` + action + `" method="post">`)
	for k, v := range param {
		b.WriteString(`<input type="hidden" name="` + k + `" value="` + v + `"/>`)
	}
	b.WriteString(`</form>`)
	b.WriteString(`<script>document.forms[0].submit();</script>`)
	return b.String()
}
