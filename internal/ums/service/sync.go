package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/ums/dto"
)

// Sync：支付查单
func Sync(ctx context.Context, req *dto.SyncReq) (*dto.SyncResp, error) {
	_ = ctx
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	param := baseParam(req.Credential)
	// 按 Java 原样：H5 查单也用 QRPAYDEFAULT
	param["instMid"] = dto.InstMidQR
	resp := &dto.SyncResp{OutTradeNo: req.OutTradeNo}

	if isQR(req.Method) {
		param["billNo"] = req.OutTradeNo
		if d := formatBillDate(req.BillDate); d != "" {
			param["billDate"] = d
		}
		result, err := client.QueryQrOrder(param)
		if err != nil {
			return nil, err
		}
		parseQrPaySync(result, resp)
	} else {
		param["merOrderId"] = req.OutTradeNo
		result, err := client.QueryH5Order(param)
		if err != nil {
			return nil, err
		}
		parseH5PaySync(result, resp)
	}
	return resp, nil
}

func parseQrPaySync(result map[string]any, resp *dto.SyncResp) {
	status := mapStr(result, "billStatus")
	switch status {
	case "PAID", "REFUND":
		resp.TradeStatus = "SUCCESS"
		resp.TotalAmount = mapStr(result, "totalAmount")
		if payment := asMap(result["billPayment"]); payment != nil {
			resp.RealAmount = mapStr(payment, "buyerPayAmount")
			resp.PayTime = mapStr(payment, "payTime")
			resp.BuyerID = mapStr(payment, "buyerId")
			resp.TargetSys = mapStr(payment, "targetSys")
			resp.TargetOrderID = mapStr(payment, "targetOrderId")
		}
	case "CLOSED":
		resp.TradeStatus = "CLOSED"
	default:
		resp.TradeStatus = "PROGRESS"
	}
}

func parseH5PaySync(result map[string]any, resp *dto.SyncResp) {
	status := mapStr(result, "status")
	switch status {
	case "TRADE_SUCCESS":
		resp.TradeStatus = "SUCCESS"
		resp.TotalAmount = mapStr(result, "totalAmount")
		resp.RealAmount = mapStr(result, "buyerPayAmount")
		resp.PayTime = mapStr(result, "payTime")
		resp.BuyerID = mapStr(result, "buyerId")
		resp.TargetSys = mapStr(result, "targetSys")
		resp.TargetOrderID = mapStr(result, "targetOrderId")
	case "TRADE_CLOSED":
		resp.TradeStatus = "CLOSED"
	default:
		resp.TradeStatus = "PROGRESS"
	}
}
