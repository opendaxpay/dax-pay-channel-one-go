package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/dto"
)

const (
	tradeClosed         = "TRADE_CLOSED"
	acqTradeStatusError = "ACQ.TRADE_STATUS_ERROR"
	acqTradeNotExist    = "ACQ.TRADE_NOT_EXIST"
)

// Close：关单或撤销
func Close(ctx context.Context, req *dto.CloseReq) (*dto.CloseResp, error) {
	if req.UseCancel {
		return doCloseOrCancel(ctx, req, "alipay.trade.cancel", "channel.error.alipayOrderCancelFailed")
	}
	return doCloseOrCancel(ctx, req, "alipay.trade.close", "channel.error.alipayCloseOrderFailed")
}

func doCloseOrCancel(ctx context.Context, req *dto.CloseReq, method, errKey string) (*dto.CloseResp, error) {
	biz := map[string]any{"out_trade_no": req.OutTradeNo}
	if req.TradeNo != "" {
		biz["trade_no"] = req.TradeNo
	}
	var out struct {
		gatewayBiz
	}
	if err := executeJSON(ctx, req.Credential, method, biz, "", &out); err != nil {
		return nil, wrapOr(err, errKey)
	}
	if out.success() {
		return closeResp(req, out.Code, out.SubCode, out.SubMsg), nil
	}
	return handleCloseFailure(ctx, req, out.SubCode, out.SubMsg, errKey)
}

func handleCloseFailure(ctx context.Context, req *dto.CloseReq, subCode, subMsg, errKey string) (*dto.CloseResp, error) {
	if subCode == acqTradeStatusError {
		syncResp, err := Sync(ctx, &dto.SyncReq{
			OutTradeNo: req.OutTradeNo,
			TradeNo:    req.TradeNo,
			Credential: req.Credential,
		})
		if err != nil {
			return nil, wrapOr(err, errKey)
		}
		if syncResp.TradeStatus == tradeClosed {
			return closeResp(req, syncResp.Code, subCode, subMsg), nil
		}
		detail := subMsg
		if detail == "" {
			detail = "交易状态不支持关闭"
		}
		return nil, alipay.NewSDKError(errKey, detail)
	}
	if subCode == acqTradeNotExist {
		return closeResp(req, "", subCode, subMsg), nil
	}
	detail := subMsg
	if detail == "" {
		detail = "关闭订单失败"
	}
	return nil, alipay.NewSDKError(errKey, detail)
}

func closeResp(req *dto.CloseReq, code, subCode, subMsg string) *dto.CloseResp {
	return &dto.CloseResp{
		OutTradeNo: req.OutTradeNo,
		TradeNo:    req.TradeNo,
		Code:       code,
		SubCode:    subCode,
		SubMsg:     subMsg,
	}
}
