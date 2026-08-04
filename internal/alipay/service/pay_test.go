package service

import (
	"testing"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/errcode"
)

// TestVerifyFailureResultUnknown：ACQ.TRADE_HAS_SUCCESS → 结果未知 10009
func TestVerifyFailureResultUnknown(t *testing.T) {
	err := verifyFailure(gatewayBiz{Code: "40004", SubCode: subCodeTradeExists, SubMsg: "交易已存在"})
	be, ok := err.(*alipay.BizError)
	if !ok {
		t.Fatalf("want *BizError, got %T", err)
	}
	if be.Code != errcode.ResultUnknown.Code {
		t.Fatalf("ACQ.TRADE_HAS_SUCCESS should map to %d, got %d", errcode.ResultUnknown.Code, be.Code)
	}
	if be.MessageKey != "channel.error.alipayPayResultUnknown" {
		t.Fatalf("unexpected messageKey: %s", be.MessageKey)
	}
}

// TestVerifyFailureSDKCallFailed：其它 subCode → SDK 调用失败 10003
func TestVerifyFailureSDKCallFailed(t *testing.T) {
	err := verifyFailure(gatewayBiz{Code: "40004", SubCode: "ACQ.TRADE_HAS_ERROR", SubMsg: "其它错误"})
	be, ok := err.(*alipay.BizError)
	if !ok {
		t.Fatalf("want *BizError, got %T", err)
	}
	if be.Code != errcode.SDKCallFailed.Code {
		t.Fatalf("other subCode should map to %d, got %d", errcode.SDKCallFailed.Code, be.Code)
	}
	if be.MessageKey != "channel.error.alipayPayCallFailed" {
		t.Fatalf("unexpected messageKey: %s", be.MessageKey)
	}
}

// TestGatewayBizSuccess：code=10000 判成功
func TestGatewayBizSuccess(t *testing.T) {
	ok := gatewayBiz{Code: codeSuccess}.success()
	if !ok {
		t.Fatal("code 10000 should be success")
	}
	inProcess := gatewayBiz{Code: codeInProcess}.success()
	if inProcess {
		t.Fatal("code 10003 should not be success")
	}
}
