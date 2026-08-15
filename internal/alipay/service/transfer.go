package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay"
	"daxpay.open/dax-pay-channel-one-go/internal/alipay/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// 支付宝转账/分账服务(alipay.fund.trans.* / alipay.trade.order.settle.*)
//
// 业务失败码(如收款人信息错误)不会触发网关异常, 通过网关公共 code 判断,
// 失败时透传 code/subCode/subMsg 由主应用决策; 仅网络/解析类异常抛 SDK 调用失败。

const (
	transferProductCode = "TRANS_ACCOUNT_NO_PWD"
	transferBizScene    = "DIRECT_TRANSFER"
	royaltyModeAsync    = "async"
)

// Transfer：发起单笔转账(alipay.fund.trans.uni.transfer)
func Transfer(ctx context.Context, req *dto.TransferReq) (*dto.TransferResp, error) {
	middleware.LoggerWithTrace(ctx).Info("alipay transfer", "outBizNo", req.OutBizNo, "amount", req.Amount)
	biz := map[string]any{
		"out_biz_no":   req.OutBizNo,
		"trans_amount": alipay.FenToYuan(int64(req.Amount)),
		"product_code": transferProductCode,
		"biz_scene":    transferBizScene,
	}
	if req.Title != "" {
		biz["order_title"] = truncateRunes(req.Title, 128)
	}
	// 收款人信息: name 必须放在 payee_info 内部, 顶层 payee_name 支付宝不识别
	payeeInfo := map[string]any{
		"identity":      req.PayeeAccount,
		"identity_type": mapIdentityType(req.PayeeType),
	}
	if req.PayeeName != "" {
		payeeInfo["name"] = req.PayeeName
	}
	biz["payee_info"] = payeeInfo
	if req.Remark != "" {
		biz["remark"] = truncateRunes(req.Remark, 200)
	}
	// 转账场景: 2026 年起新接入商户必填
	if req.TransferSceneName != "" {
		biz["transfer_scene_name"] = req.TransferSceneName
	}
	if len(req.ReportInfos) > 0 {
		biz["transfer_scene_report_infos"] = buildReportInfos(req.ReportInfos)
	}
	cleanBiz(biz)

	var out struct {
		gatewayBiz
		OrderID        string `json:"order_id"`
		Status         string `json:"status"`
		PayFundOrderID string `json:"pay_fund_order_id"`
		TransDate      string `json:"trans_date"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.fund.trans.uni.transfer", biz, req.NotifyURL, &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayTransferFailed")
	}
	resp := &dto.TransferResp{Code: out.Code, SubCode: out.SubCode, SubMsg: out.SubMsg}
	// 业务失败时业务字段均为空, 直接返回让主应用决策
	if !out.success() {
		middleware.LoggerWithTrace(ctx).Warn("alipay transfer biz failed",
			"outBizNo", req.OutBizNo, "code", out.Code, "subCode", out.SubCode, "subMsg", out.SubMsg)
		return resp, nil
	}
	resp.OrderID = out.OrderID
	resp.Status = out.Status
	resp.FailReason = out.SubMsg
	resp.PayFundOrderID = out.PayFundOrderID
	resp.TransDate = out.TransDate
	return resp, nil
}

// TransferSync：同步查询转账状态(alipay.fund.trans.common.query)
func TransferSync(ctx context.Context, req *dto.TransferReq) (*dto.TransferResp, error) {
	biz := map[string]any{
		"product_code": transferProductCode,
		"biz_scene":    transferBizScene,
		"out_biz_no":   req.OutBizNo,
	}
	var out struct {
		gatewayBiz
		OrderID        string `json:"order_id"`
		Status         string `json:"status"`
		FailReason     string `json:"fail_reason"`
		PayDate        string `json:"pay_date"`
		PayFundOrderID string `json:"pay_fund_order_id"`
		ErrorCode      string `json:"error_code"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.fund.trans.common.query", biz, "", &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayTransferQueryFailed")
	}
	resp := &dto.TransferResp{Code: out.Code, SubCode: out.SubCode, SubMsg: out.SubMsg}
	// 业务失败时业务字段均为空, 直接返回让主应用按 subCode 决策资金态
	if !out.success() {
		middleware.LoggerWithTrace(ctx).Warn("alipay transfer sync biz failed",
			"outBizNo", req.OutBizNo, "code", out.Code, "subCode", out.SubCode, "subMsg", out.SubMsg)
		return resp, nil
	}
	resp.OrderID = out.OrderID
	resp.Status = out.Status
	resp.FailReason = out.FailReason
	resp.FinishTime = out.PayDate
	resp.PayFundOrderID = out.PayFundOrderID
	resp.ErrorCode = out.ErrorCode
	return resp, nil
}

// Alloc：发起分账(alipay.trade.order.settle)
func Alloc(ctx context.Context, req *dto.AllocReq) (*dto.AllocResp, error) {
	middleware.LoggerWithTrace(ctx).Info("alipay alloc", "outRequestNo", req.OutRequestNo, "tradeNo", req.TradeNo)
	royaltyMode := req.RoyaltyMode
	if royaltyMode == "" {
		royaltyMode = royaltyModeAsync
	}
	biz := map[string]any{
		"out_request_no": req.OutRequestNo,
		"trade_no":       req.TradeNo,
		"royalty_mode":   royaltyMode,
	}
	// 分账子参数
	royaltyParams := make([]map[string]any, 0, len(req.RoyaltyParameters))
	for _, rp := range req.RoyaltyParameters {
		item := map[string]any{
			"trans_in": rp.TransIn,
			"amount":   alipay.FenToYuan(int64(rp.Amount)),
		}
		if rp.TransInType != "" {
			item["trans_in_type"] = rp.TransInType
		}
		royaltyParams = append(royaltyParams, item)
	}
	biz["royalty_parameters"] = royaltyParams

	var out struct {
		gatewayBiz
		SettleNo string `json:"settle_no"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.order.settle", biz, "", &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayAllocFailed")
	}
	resp := &dto.AllocResp{Code: out.Code, SubCode: out.SubCode, SubMsg: out.SubMsg}
	// 业务失败时业务字段均为空, 直接返回让主应用决策
	if !out.success() {
		middleware.LoggerWithTrace(ctx).Warn("alipay alloc biz failed",
			"outRequestNo", req.OutRequestNo, "code", out.Code, "subCode", out.SubCode, "subMsg", out.SubMsg)
		return resp, nil
	}
	resp.SettleNo = out.SettleNo
	return resp, nil
}

// AllocSync：分账同步查询(alipay.trade.order.settle.query)
func AllocSync(ctx context.Context, req *dto.AllocReq) (*dto.AllocResp, error) {
	biz := map[string]any{
		"out_request_no": req.OutRequestNo,
		"trade_no":       req.TradeNo,
	}
	var out struct {
		gatewayBiz
		SettleNo          string `json:"settle_no"`
		RoyaltyDetailList []struct {
			DetailID  string `json:"detail_id"`
			TransIn   string `json:"trans_in"`
			State     string `json:"state"`
			ErrorDesc string `json:"error_desc"`
			ExecuteDt string `json:"execute_dt"`
		} `json:"royalty_detail_list"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.order.settle.query", biz, "", &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayAllocQueryFailed")
	}
	resp := &dto.AllocResp{Code: out.Code, SubCode: out.SubCode, SubMsg: out.SubMsg}
	if !out.success() {
		middleware.LoggerWithTrace(ctx).Warn("alipay alloc sync biz failed",
			"outRequestNo", req.OutRequestNo, "code", out.Code, "subCode", out.SubCode, "subMsg", out.SubMsg)
		return resp, nil
	}
	// 查询响应不返回 settleNo, outRequestNo 即平台 allocNo
	resp.SettleNo = req.OutRequestNo
	// 映射逐明细结果
	for _, r := range out.RoyaltyDetailList {
		resp.RoyaltyDetailList = append(resp.RoyaltyDetailList, dto.RoyaltyDetailResult{
			DetailID:  r.DetailID,
			TransIn:   r.TransIn,
			State:     r.State,
			ErrorDesc: r.ErrorDesc,
			ExecuteDt: r.ExecuteDt,
		})
	}
	return resp, nil
}

// mapIdentityType：平台收款人类型 → 支付宝身份类型
//
// user_id → ALIPAY_USER_ID / open_id → ALIPAY_OPEN_ID / login_name → ALIPAY_LOGON_ID
func mapIdentityType(payeeType string) string {
	switch payeeType {
	case "open_id":
		return "ALIPAY_OPEN_ID"
	case "login_name":
		return "ALIPAY_LOGON_ID"
	default:
		return "ALIPAY_USER_ID"
	}
}

func buildReportInfos(infos []dto.ReportInfo) []map[string]any {
	result := make([]map[string]any, 0, len(infos))
	for _, info := range infos {
		result = append(result, map[string]any{
			"info_type":    info.InfoType,
			"info_content": info.InfoContent,
		})
	}
	return result
}

// truncateRunes：按字符截断(UTF-8 安全, 对齐 Java StrUtil.sub)
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
