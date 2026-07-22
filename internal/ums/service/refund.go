package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/ums/dto"
)

// Refund：退款
func Refund(ctx context.Context, req *dto.RefundReq) (*dto.RefundResp, error) {
	_ = ctx
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	param := baseParam(req.Credential)
	param["refundOrderId"] = req.OutRefundNo
	param["refundAmount"] = int64(req.RefundAmount)
	if req.Reason != "" {
		param["refundDesc"] = req.Reason
	}

	var result map[string]any
	var finishField string
	if isQR(req.Method) {
		param["billNo"] = req.OutTradeNo
		if d := formatBillDate(req.BillDate); d != "" {
			param["billDate"] = d
		}
		param["instMid"] = dto.InstMidQR
		result, err = client.RefundQr(param)
		finishField = "refundPayTime"
	} else {
		param["merOrderId"] = req.OutTradeNo
		param["instMid"] = dto.InstMidH5
		result, err = client.RefundH5(param)
		finishField = "payTime"
	}
	if err != nil {
		return nil, err
	}
	return &dto.RefundResp{
		OutRefundNo:  req.OutRefundNo,
		RefundStatus: mapStr(result, "refundStatus"),
		FinishTime:   mapStr(result, finishField),
	}, nil
}

// RefundSync：退款查询
func RefundSync(ctx context.Context, req *dto.RefundSyncReq) (*dto.RefundSyncResp, error) {
	_ = ctx
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	param := baseParam(req.Credential)
	resp := &dto.RefundSyncResp{OutRefundNo: req.OutRefundNo}

	if isQR(req.Method) {
		param["instMid"] = dto.InstMidQR
		param["billNo"] = req.OutTradeNo
		param["refundOrderId"] = req.OutRefundNo
		if d := formatBillDate(req.BillDate); d != "" {
			param["billDate"] = d
		}
		result, err := client.QueryQrOrder(param)
		if err != nil {
			return nil, err
		}
		parseQrRefundSync(result, resp)
	} else {
		param["instMid"] = dto.InstMidH5
		param["merOrderId"] = req.OutRefundNo
		result, err := client.QueryH5Refund(param)
		if err != nil {
			return nil, err
		}
		parseH5RefundSync(result, resp)
	}
	return resp, nil
}

func parseQrRefundSync(result map[string]any, resp *dto.RefundSyncResp) {
	payment := asMap(result["refundBillPayment"])
	if payment == nil {
		resp.RefundStatus = "PROGRESS"
		return
	}
	status := mapStr(payment, "status")
	resp.RefundAmount = mapStr(payment, "totalAmount")
	resp.FinishTime = mapStr(payment, "payTime")
	switch status {
	case "TRADE_SUCCESS":
		resp.RefundStatus = "SUCCESS"
	case "TRADE_REFUND":
		resp.RefundStatus = "CLOSED"
	default:
		resp.RefundStatus = "PROGRESS"
	}
}

func parseH5RefundSync(result map[string]any, resp *dto.RefundSyncResp) {
	status := mapStr(result, "refundStatus")
	switch status {
	case "SUCCESS":
		resp.RefundStatus = "SUCCESS"
		resp.RefundAmount = mapStr(result, "totalAmount")
		resp.FinishTime = mapStr(result, "payTime")
	case "FAIL":
		resp.RefundStatus = "CLOSED"
	default:
		resp.RefundStatus = "PROGRESS"
	}
}
