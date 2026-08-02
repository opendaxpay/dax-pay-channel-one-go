package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/union/dto"
)

// Sync 云闪付支付同步(通过银联 queryTrans.do 查询订单状态)
func Sync(ctx context.Context, req *dto.SyncReq) (*dto.SyncResp, error) {
	_ = ctx
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	param := baseParam(req.Credential)
	param["orderId"] = req.OutTradeNo
	// 银联查询交易类型
	param["txnType"] = "00"
	param["txnSubType"] = "00"
	param["bizType"] = "000000"
	result, err := client.QueryTrans(param)
	if err != nil {
		return nil, err
	}
	return parseSyncResp(req.OutTradeNo, result), nil
}

// parseSyncResp：解析查询响应(respCode 查询结果 + origRespCode 原交易结果)
func parseSyncResp(outTradeNo string, resp map[string]string) *dto.SyncResp {
	r := &dto.SyncResp{OutTradeNo: outTradeNo}
	if resp["respCode"] != "00" {
		// 查询本身失败, 订单状态未知, 视为进行中
		r.TradeStatus = "PROGRESS"
		r.ErrorMsg = resp["respMsg"]
		return r
	}
	r.TradeStatus = mapPayStatus(resp["origRespCode"])
	r.TotalAmount = resp["txnAmt"]
	r.PayTime = resp["txnTime"]
	// queryId 作为通道订单号(退款时作为 origQryId, 由主应用存入 trade.outOrderNo)
	r.QueryID = resp["queryId"]
	r.BuyerID = resp["accNo"]
	return r
}

// mapPayStatus：origRespCode 00→SUCCESS, 05→CLOSED, 其他→PROGRESS
func mapPayStatus(origRespCode string) string {
	switch origRespCode {
	case "00":
		return "SUCCESS"
	case "05":
		return "CLOSED"
	default:
		return "PROGRESS"
	}
}
