package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/dto"
)

// 本文件测试分账接收方绑定/解绑(alipay.trade.royalty.relation.bind / unbind)，
// 通过 httptest 假网关覆盖：类型映射、name 携带规则、业务失败透传、USER_NOT_EXIST 幂等容错。
// 假网关响应不带 sign 字段，网关客户端按约定跳过验签。

// newReceiverTestCredential：生成测试凭证(自签 RSA 密钥对 + ServerURL 指向假网关)
func newReceiverTestCredential(t *testing.T, handler http.HandlerFunc) *alipay.SdkCredential {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return &alipay.SdkCredential{
		AliAppId:        "2021000000000001",
		PrivateKey:      string(privPEM),
		AlipayPublicKey: string(pubPEM),
		ServerURL:       ts.URL,
		SignType:        "RSA2",
	}
}

// fakeGateway：按调用顺序返回预设响应体，并捕获请求参数
func fakeGateway(t *testing.T, responses []string, captured map[string]string) http.HandlerFunc {
	idx := 0
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		for k, v := range r.PostForm {
			captured[k] = v[0]
		}
		w.Header().Set("Content-Type", "application/json")
		if idx >= len(responses) {
			t.Fatalf("unexpected gateway call #%d", idx+1)
		}
		_, _ = w.Write([]byte(responses[idx]))
		idx++
	}
}

// parseBizContent：从捕获的请求参数解析 biz_content
func parseBizContent(t *testing.T, captured map[string]string) map[string]any {
	t.Helper()
	raw := captured["biz_content"]
	if raw == "" {
		t.Fatal("missing biz_content in gateway request")
	}
	var biz map[string]any
	if err := json.Unmarshal([]byte(raw), &biz); err != nil {
		t.Fatalf("parse biz_content: %v", err)
	}
	return biz
}

// receiverReq：构造测试请求
func receiverReq(cred *alipay.SdkCredential, receiverName string) *dto.AllocReceiverReq {
	return &dto.AllocReceiverReq{
		OutRequestNo:    "alloc-recv-001",
		ReceiverType:    "USER_ID",
		ReceiverAccount: "2088000000000001",
		ReceiverName:    receiverName,
		Credential:      cred,
	}
}

// TestToAlipayReceiverType：平台接收方类型 → 支付宝原生小写类型
func TestToAlipayReceiverType(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"USER_ID", "userId"},
		{"LOGIN_NAME", "loginName"},
		{"未知类型原样透传", "未知类型原样透传"},
		{"", ""},
	}
	for _, c := range cases {
		if got := toAlipayReceiverType(c.in); got != c.want {
			t.Fatalf("toAlipayReceiverType(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestBindReceiver：绑定场景（成功 / 业务失败透传 / name 携带规则）
func TestBindReceiver(t *testing.T) {
	captured := map[string]string{}
	// 第 1 次成功、第 2 次业务失败
	cred := newReceiverTestCredential(t, fakeGateway(t, []string{
		`{"alipay_trade_royalty_relation_bind_response":{"code":"10000","msg":"Success"}}`,
		`{"alipay_trade_royalty_relation_bind_response":{"code":"40004","msg":"Business Failed","sub_code":"ISV_INVALID_RECEIVER","sub_msg":"接收方账号无效"}}`,
	}, captured))

	// 场景 1：成功 + 名称非空 → 响应无错误字段, 请求携带 name 且类型已映射小写
	resp, err := BindReceiver(context.Background(), receiverReq(cred, "测试接收方"))
	if err != nil {
		t.Fatalf("bind success case: unexpected err %v", err)
	}
	if resp.ErrorCode != "" || resp.ErrorMsg != "" {
		t.Fatalf("bind success case: expect empty error fields, got %+v", resp)
	}
	biz := parseBizContent(t, captured)
	if biz["out_request_no"] != "alloc-recv-001" {
		t.Fatalf("out_request_no = %v, want alloc-recv-001", biz["out_request_no"])
	}
	list, ok := biz["receiver_list"].([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("receiver_list = %v, want single element", biz["receiver_list"])
	}
	item, _ := list[0].(map[string]any)
	if item["type"] != "userId" {
		t.Fatalf("receiver type = %v, want userId(大写枚举映射小写)", item["type"])
	}
	if item["account"] != "2088000000000001" {
		t.Fatalf("receiver account = %v", item["account"])
	}
	if item["name"] != "测试接收方" {
		t.Fatalf("receiver name = %v, want 测试接收方(绑定时携带)", item["name"])
	}

	// 场景 2：业务失败 → subCode/subMsg 透传进 errorCode/errorMsg, 不抛错
	resp, err = BindReceiver(context.Background(), receiverReq(cred, ""))
	if err != nil {
		t.Fatalf("bind biz-failed case: unexpected err %v", err)
	}
	if resp.ErrorCode != "ISV_INVALID_RECEIVER" || resp.ErrorMsg != "接收方账号无效" {
		t.Fatalf("bind biz-failed case: got %+v, want subCode/subMsg 透传", resp)
	}
	// 名称留空 → 请求不携带 name 字段
	if strings.Contains(captured["biz_content"], "\"name\"") {
		t.Fatalf("biz_content should not contain name when receiverName empty: %s", captured["biz_content"])
	}
}

// TestUnbindReceiver：解绑场景（成功 / USER_NOT_EXIST 幂等 / 其他业务失败透传）
func TestUnbindReceiver(t *testing.T) {
	captured := map[string]string{}
	// 第 1 次成功、第 2 次分账方不存在、第 3 次其他业务失败
	cred := newReceiverTestCredential(t, fakeGateway(t, []string{
		`{"alipay_trade_royalty_relation_unbind_response":{"code":"10000","msg":"Success"}}`,
		`{"alipay_trade_royalty_relation_unbind_response":{"code":"40004","msg":"Business Failed","sub_code":"USER_NOT_EXIST","sub_msg":"分账关系不存在"}}`,
		`{"alipay_trade_royalty_relation_unbind_response":{"code":"40004","msg":"Business Failed","sub_code":"ISV_INVALID_RECEIVER","sub_msg":"接收方账号无效"}}`,
	}, captured))

	// 场景 1：成功 → 响应无错误字段, 请求不含 name(解绑仅类型+账号)
	resp, err := UnbindReceiver(context.Background(), receiverReq(cred, "名称应被忽略"))
	if err != nil {
		t.Fatalf("unbind success case: unexpected err %v", err)
	}
	if resp.ErrorCode != "" || resp.ErrorMsg != "" {
		t.Fatalf("unbind success case: expect empty error fields, got %+v", resp)
	}
	if strings.Contains(captured["biz_content"], "\"name\"") {
		t.Fatalf("unbind biz_content should not contain name: %s", captured["biz_content"])
	}

	// 场景 2：USER_NOT_EXIST → 幂等容错视为成功(空错误字段), 对齐 Java 分支语义
	resp, err = UnbindReceiver(context.Background(), receiverReq(cred, ""))
	if err != nil {
		t.Fatalf("unbind user-not-exist case: unexpected err %v", err)
	}
	if resp.ErrorCode != "" || resp.ErrorMsg != "" {
		t.Fatalf("unbind user-not-exist case: expect idempotent success with empty error fields, got %+v", resp)
	}

	// 场景 3：其他业务失败 → 正常透传
	resp, err = UnbindReceiver(context.Background(), receiverReq(cred, ""))
	if err != nil {
		t.Fatalf("unbind biz-failed case: unexpected err %v", err)
	}
	if resp.ErrorCode != "ISV_INVALID_RECEIVER" || resp.ErrorMsg != "接收方账号无效" {
		t.Fatalf("unbind biz-failed case: got %+v, want subCode/subMsg 透传", resp)
	}
}
