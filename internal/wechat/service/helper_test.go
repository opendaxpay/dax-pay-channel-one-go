package service

import (
	"errors"
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/openapi"
)

func TestIsPayResultUnknown(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"ORDERPAID", &openapi.APIError{Code: "ORDERPAID"}, true},
		{"AUTH_CODE_USED", &openapi.APIError{Code: "AUTH_CODE_USED"}, true},
		{"USERPAYING应判false", &openapi.APIError{Code: "USERPAYING"}, false},
		{"其它错误", &openapi.APIError{Code: "SYSTEMERROR"}, false},
		{"非APIError", errors.New("boom"), false},
		{"nil", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isPayResultUnknown(c.err); got != c.want {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestIsCodepayPaying(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"USERPAYING", &openapi.APIError{Code: "USERPAYING"}, true},
		{"SYSTEMERROR", &openapi.APIError{Code: "SYSTEMERROR"}, true},
		{"ORDERPAID应判false", &openapi.APIError{Code: "ORDERPAID"}, false},
		{"非APIError", errors.New("boom"), false},
		{"nil", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isCodepayPaying(c.err); got != c.want {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}

// TestWrapAPIErrResultUnknown：订单已支付/付款码已被使用 → 结果未知 10009
func TestWrapAPIErrResultUnknown(t *testing.T) {
	for _, code := range []string{"ORDERPAID", "AUTH_CODE_USED"} {
		err := wrapAPIErr("channel.error.wechatPayCallFailed", &openapi.APIError{Code: code})
		be, ok := err.(*wechat.BizError)
		if !ok {
			t.Fatalf("%s: want *BizError, got %T", code, err)
		}
		if be.Code != errcode.ResultUnknown.Code {
			t.Fatalf("%s: want code %d, got %d", code, errcode.ResultUnknown.Code, be.Code)
		}
		if be.MessageKey != "channel.error.wechatPayResultUnknown" {
			t.Fatalf("%s: want messageKey channel.error.wechatPayResultUnknown, got %s", code, be.MessageKey)
		}
	}
}

// TestWrapAPIErrSDKCallFailed：其它错误 → SDK 调用失败 10003
func TestWrapAPIErrSDKCallFailed(t *testing.T) {
	err := wrapAPIErr("channel.error.wechatPayCallFailed", &openapi.APIError{Code: "PARAM_ERROR"})
	be, ok := err.(*wechat.BizError)
	if !ok {
		t.Fatalf("want *BizError, got %T", err)
	}
	if be.Code != errcode.SDKCallFailed.Code {
		t.Fatalf("want code %d, got %d", errcode.SDKCallFailed.Code, be.Code)
	}
}

// TestWrapAPIErrBizErrorPassthrough：已是 BizError 直接透传
func TestWrapAPIErrBizErrorPassthrough(t *testing.T) {
	orig := wechat.NewSDKError("channel.error.wechatPayCallFailed", "detail")
	err := wrapAPIErr("channel.error.wechatPayCallFailed", orig)
	if err != orig {
		t.Fatal("BizError should pass through unchanged")
	}
}
