package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/union/dto"
)

// RefundSync 云闪付退款同步(通过银联 queryTrans.do 查询退款单状态)
func RefundSync(ctx context.Context, req *dto.RefundSyncReq) (*dto.RefundSyncResp, error) {
	_ = ctx
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	param := baseParam(req.Credential)
	param["orderId"] = req.OutRefundNo
	param["txnType"] = "00"
	param["txnSubType"] = "00"
	param["bizType"] = "000000"
	result, err := client.QueryTrans(param)
	if err != nil {
		return nil, err
	}
	return parseRefundSyncResp(req.OutRefundNo, result), nil
}

func parseRefundSyncResp(outRefundNo string, resp map[string]string) *dto.RefundSyncResp {
	r := &dto.RefundSyncResp{OutRefundNo: outRefundNo}
	if resp["respCode"] != "00" {
		r.RefundStatus = "PROGRESS"
		r.ErrorMsg = resp["respMsg"]
		return r
	}
	r.RefundStatus = mapRefundSyncStatus(resp["origRespCode"])
	r.RefundAmount = resp["txnAmt"]
	r.FinishTime = resp["txnTime"]
	return r
}

// mapRefundSyncStatus：origRespCode 00→SUCCESS, 05→CLOSED, 其他→PROGRESS
func mapRefundSyncStatus(origRespCode string) string {
	switch origRespCode {
	case "00":
		return "SUCCESS"
	case "05":
		return "CLOSED"
	default:
		return "PROGRESS"
	}
}
