package service

import (
	"context"
	"fmt"

	"daxpay.open/dax-pay-channel-one-go/internal/union"
	"daxpay.open/dax-pay-channel-one-go/internal/union/dto"
)

// Refund 云闪付退款(银联退货交易类型 04, 需原交易查询凭证 origQueryId)
func Refund(ctx context.Context, req *dto.RefundReq) (*dto.RefundResp, error) {
	_ = ctx
	if req.OrigQueryID == "" {
		return nil, union.NewSDKError("channel.error.unionRequestFailed", "退款需传入原交易查询凭证 origQueryId")
	}
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	param := baseParam(req.Credential)
	param["orderId"] = req.OutRefundNo
	param["origQryId"] = req.OrigQueryID
	param["txnAmt"] = fmt.Sprintf("%d", int64(req.RefundAmount))
	param["backUrl"] = req.NotifyURL
	result, err := client.Refund(param)
	if err != nil {
		return nil, err
	}
	return &dto.RefundResp{
		OutRefundNo:  req.OutRefundNo,
		RefundStatus: mapRefundStatus(result["respCode"]),
		FinishTime:   result["txnTime"],
	}, nil
}

// mapRefundStatus：respCode 00→SUCCESS, 03→PROCESSING, 其他→FAIL
func mapRefundStatus(respCode string) string {
	switch respCode {
	case "00":
		return "SUCCESS"
	case "03":
		return "PROCESSING"
	default:
		return "FAIL"
	}
}
