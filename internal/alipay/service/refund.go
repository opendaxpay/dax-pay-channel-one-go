package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

const fundChangeY = "Y"

// Refund：退款
func Refund(ctx context.Context, req *dto.RefundReq) (*dto.RefundResp, error) {
	middleware.LoggerWithTrace(ctx).Info("alipay refund",
		"outTradeNo", req.OutTradeNo, "outRequestNo", req.OutRequestNo)

	biz := map[string]any{
		"out_trade_no":   req.OutTradeNo,
		"out_request_no": req.OutRequestNo,
		"refund_amount":  alipay.FenToYuan(int64(req.RefundAmount)),
		"query_options":  []string{"deposit_back_info"},
	}
	if req.TradeNo != "" {
		biz["trade_no"] = req.TradeNo
	}

	resp := &dto.RefundResp{
		OutTradeNo:   req.OutTradeNo,
		TradeNo:      req.TradeNo,
		OutRequestNo: req.OutRequestNo,
		Complete:     false,
	}

	var out struct {
		gatewayBiz
		OutTradeNo   string `json:"out_trade_no"`
		TradeNo      string `json:"trade_no"`
		FundChange   string `json:"fund_change"`
		GmtRefundPay string `json:"gmt_refund_pay"`
		BuyerUserID  string `json:"buyer_user_id"`
		BuyerOpenID  string `json:"buyer_open_id"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.refund", biz, "", &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayRefundCallFailed")
	}
	if !out.success() {
		return nil, alipay.NewSDKError("channel.error.alipayRefundCallFailed", out.errDetail())
	}
	resp.OutTradeNo = out.OutTradeNo
	resp.TradeNo = out.TradeNo
	resp.FundChange = out.FundChange
	resp.BuyerUserID = out.BuyerUserID
	resp.BuyerOpenID = out.BuyerOpenID
	if out.FundChange == fundChangeY {
		resp.Complete = true
		resp.FinishTime = alipay.ParseGatewayTime(out.GmtRefundPay)
	}
	return resp, nil
}

// RefundSync：退款查询
func RefundSync(ctx context.Context, req *dto.RefundSyncReq) (*dto.RefundSyncResp, error) {
	middleware.LoggerWithTrace(ctx).Info("alipay refund sync",
		"outTradeNo", req.OutTradeNo, "outRequestNo", req.OutRequestNo)

	biz := map[string]any{
		"out_trade_no":   req.OutTradeNo,
		"out_request_no": req.OutRequestNo,
	}
	if req.TradeNo != "" {
		biz["trade_no"] = req.TradeNo
	}

	var out struct {
		gatewayBiz
		RefundStatus string `json:"refund_status"`
		OutTradeNo   string `json:"out_trade_no"`
		TradeNo      string `json:"trade_no"`
		OutRequestNo string `json:"out_request_no"`
		GmtRefundPay string `json:"gmt_refund_pay"`
		RefundAmount string `json:"refund_amount"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.fastpay.refund.query", biz, "", &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayRefundQueryFailed")
	}
	return &dto.RefundSyncResp{
		Code:         out.Code,
		SubCode:      out.SubCode,
		SubMsg:       out.SubMsg,
		RefundStatus: out.RefundStatus,
		OutTradeNo:   out.OutTradeNo,
		TradeNo:      out.TradeNo,
		OutRequestNo: out.OutRequestNo,
		FinishTime:   alipay.ParseGatewayTime(out.GmtRefundPay),
		RefundAmount: alipay.YuanToFenPtr(out.RefundAmount),
	}, nil
}
