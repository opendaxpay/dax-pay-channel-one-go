package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/openapi"
)

// 微信转账/分账服务(商家转账到零钱 V3 /v3/fund-app/mch-transfer/* 与 profitsharing/*)
//
// 转账: SDK 级异常抛 SDK_CALL_FAILED; 分账: 异常不抛, 包 resp 返回 errorCode/errorMsg 由主应用决策(对齐 Java)。
// 注意: V3 API 路径统一带 /v3/ 前缀(转账曾漏写前缀, 网关对裸路径返回 nginx 404 页)。

const (
	transferBillsPath          = "/v3/fund-app/mch-transfer/transfer-bills"
	transferBillNoQueryPath    = "/v3/fund-app/mch-transfer/transfer-bills/transfer-bill-no/%s"
	transferOutBillNoQueryPath = "/v3/fund-app/mch-transfer/transfer-bills/out-bill-no/%s"
	profitSharingOrdersPath    = "/v3/profitsharing/orders"
	profitSharingQueryPath     = "/v3/profitsharing/orders/%s"
)

// Transfer：发起商家转账到零钱(V3 /fund-app/mch-transfer/transfer-bills)
//
// user_name 为敏感字段, 用微信支付平台证书公钥 RSA-OAEP(SHA-1) 加密上送。
func Transfer(ctx context.Context, req *dto.TransferReq) (*dto.TransferResp, error) {
	middleware.LoggerWithTrace(ctx).Info("wechat transfer", "outBillNo", req.OutBillNo, "amount", req.Amount)
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"appid":             req.Credential.WxAppId,
		"out_bill_no":       req.OutBillNo,
		"transfer_scene_id": req.Scene,
		"openid":            req.Openid,
		"transfer_amount":   int64(req.Amount),
		"transfer_remark":   truncateRunes(req.Remark, 32),
	}
	if req.NotifyURL != "" {
		body["notify_url"] = req.NotifyURL
	}
	if req.UserName != "" {
		// 敏感字段: 微信支付平台证书 RSA-OAEP 加密
		pub, err := client.PlatformCert(ctx)
		if err != nil {
			return nil, wechat.NewSDKError("channel.error.wechatTransferFailed", err.Error())
		}
		encrypted, err := openapi.EncryptOAEP(pub, req.UserName)
		if err != nil {
			return nil, wechat.NewSDKError("channel.error.wechatTransferFailed", err.Error())
		}
		body["user_name"] = encrypted
	}
	// 转账场景报备信息(微信接口必填): infoContent 留空用 `-` 兜底
	if len(req.ReportInfos) > 0 {
		infos := make([]map[string]any, 0, len(req.ReportInfos))
		for _, info := range req.ReportInfos {
			content := info.InfoContent
			if content == "" {
				content = "-"
			}
			infos = append(infos, map[string]any{
				"info_type":    info.InfoType,
				"info_content": content,
			})
		}
		body["transfer_scene_report_infos"] = infos
	}

	raw, err := client.Do(ctx, http.MethodPost, transferBillsPath, body)
	if err != nil {
		return nil, wechat.NewSDKError("channel.error.wechatTransferFailed", err.Error())
	}
	var out struct {
		TransferBillNo string `json:"transfer_bill_no"`
		PackageInfo    string `json:"package_info"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, wechat.NewSDKError("channel.error.wechatTransferFailed", err.Error())
	}
	// package_info: 需商户二次确认时用于拉起确认页
	return &dto.TransferResp{TransferBillNo: out.TransferBillNo, PackageInfo: out.PackageInfo}, nil
}

// TransferSync：同步查询转账状态
//
// 优先按通道转账单号查询, 缺失或与平台单号相同时按商户单号查询(对齐 Java 分支逻辑)。
func TransferSync(ctx context.Context, req *dto.TransferReq) (*dto.TransferResp, error) {
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	var path string
	if req.OutBillNo != "" && req.OutBillNo != req.TransferNo {
		path = fmt.Sprintf(transferBillNoQueryPath, req.OutBillNo)
	} else {
		path = fmt.Sprintf(transferOutBillNoQueryPath, req.TransferNo)
	}
	raw, err := client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, wechat.NewSDKError("channel.error.wechatTransferQueryFailed", err.Error())
	}
	var out struct {
		TransferBillNo string `json:"transfer_bill_no"`
		State          string `json:"state"`
		FailReason     string `json:"fail_reason"`
		UpdateTime     string `json:"update_time"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, wechat.NewSDKError("channel.error.wechatTransferQueryFailed", err.Error())
	}
	resp := &dto.TransferResp{
		TransferBillNo: out.TransferBillNo,
		State:          out.State,
		FailReason:     out.FailReason,
		FinishTime:     parseTransferTime(out.UpdateTime),
		Complete:       isTransferTerminal(out.State),
	}
	return resp, nil
}

// Alloc：发起分账(V3 profitsharing/orders, unfreeze_unsplit=true 自动解冻剩余)
func Alloc(ctx context.Context, req *dto.AllocReq) (*dto.AllocResp, error) {
	middleware.LoggerWithTrace(ctx).Info("wechat alloc", "outOrderNo", req.OutOrderNo, "transactionId", req.TransactionID)
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	// 接收方列表: description 留空补 "订单分账"
	receivers := make([]map[string]any, 0, len(req.Receivers))
	for _, r := range req.Receivers {
		item := map[string]any{
			"type":    r.Type,
			"account": r.Account,
			"amount":  int64(r.Amount),
		}
		if r.Name != "" {
			item["name"] = r.Name
		}
		description := r.Description
		if description == "" {
			description = "订单分账"
		}
		item["description"] = description
		receivers = append(receivers, item)
	}
	body := map[string]any{
		"transaction_id":   req.TransactionID,
		"out_order_no":     req.OutOrderNo,
		"unfreeze_unsplit": true,
		"receivers":        receivers,
	}
	raw, err := client.Do(ctx, http.MethodPost, profitSharingOrdersPath, body)
	if err != nil {
		// 异常不抛, 包 resp 返回 errorCode/errorMsg 由主应用决策
		return allocErrorResp(err), nil
	}
	var out struct {
		TransactionID string `json:"transaction_id"`
		State         string `json:"state"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return &dto.AllocResp{ErrorMsg: err.Error()}, nil
	}
	return &dto.AllocResp{TransactionID: out.TransactionID, State: out.State}, nil
}

// AllocSync：分账同步查询(V3 profitsharing/orders/{out_order_no})
func AllocSync(ctx context.Context, req *dto.AllocReq) (*dto.AllocResp, error) {
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf(profitSharingQueryPath, req.OutOrderNo) + "?transaction_id=" + url.QueryEscape(req.TransactionID)
	raw, err := client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		// 异常不抛, 包 resp 返回 errorCode/errorMsg 由主应用决策
		return allocErrorResp(err), nil
	}
	var out struct {
		TransactionID string `json:"transaction_id"`
		State         string `json:"state"`
		Receivers     []struct {
			Account    string `json:"account"`
			Amount     *int64 `json:"amount"`
			Result     string `json:"result"`
			FailReason string `json:"fail_reason"`
			FinishTime string `json:"finish_time"`
		} `json:"receivers"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return &dto.AllocResp{ErrorMsg: err.Error()}, nil
	}
	resp := &dto.AllocResp{TransactionID: out.TransactionID, State: out.State}
	// 映射逐明细结果
	for _, r := range out.Receivers {
		rr := dto.ReceiverResult{
			Account:    r.Account,
			Result:     r.Result,
			FailReason: r.FailReason,
			FinishTime: parseAllocTime(r.FinishTime),
		}
		if r.Amount != nil {
			rr.Amount = ptrInt64(*r.Amount)
		}
		resp.Receivers = append(resp.Receivers, rr)
	}
	return resp, nil
}

// allocErrorResp：分账异常 → 包 resp(对齐 Java 不抛异常, 透传 errorCode/errorMsg)
func allocErrorResp(err error) *dto.AllocResp {
	resp := &dto.AllocResp{}
	if apiErr, ok := err.(*openapi.APIError); ok {
		resp.ErrorCode = apiErr.Code
		resp.ErrorMsg = apiErr.Message
		if resp.ErrorMsg == "" {
			resp.ErrorMsg = apiErr.Error()
		}
		return resp
	}
	resp.ErrorMsg = err.Error()
	return resp
}
