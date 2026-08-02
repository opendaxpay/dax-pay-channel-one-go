package service

import (
	"daxpay.open/dax-pay-channel-one-go/internal/union"
	"daxpay.open/dax-pay-channel-one-go/internal/union/openapi"
)

func newClient(cred *union.SdkCredential) (*openapi.Client, error) {
	if cred == nil {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", "credential is required")
	}
	return openapi.NewClient(cred), nil
}

// baseParam：银联公共参数(version/encoding/merId/txnTime/accessType/currencyCode)
func baseParam(cred *union.SdkCredential) map[string]string {
	return map[string]string{
		"version":      "5.1.0",
		"encoding":     "UTF-8",
		"merId":        cred.MerID,
		"txnTime":      openapi.TxnTime(),
		"accessType":   "0",
		"currencyCode": "156",
	}
}
