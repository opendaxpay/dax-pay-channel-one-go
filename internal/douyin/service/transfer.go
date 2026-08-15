package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin"
	"daxpay.open/dax-pay-channel-one-go/internal/douyin/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/douyin/openapi"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// 抖音转账/分账服务
//
// 转账: 商家转账原生 API(/v1/fund_trade/mch-transfer/*), 收款人姓名/手机号为敏感字段,
// 用平台证书 RSA(PKCS#1 v1.5) 加密上送并携带 Douyinpay-Serial 头, SDK 级异常抛 SDK_CALL_FAILED。
// 分账: /v1/trade/profitsharing/*, 接收方姓名同样加密; 异常不抛, 包 resp 返回 errorCode/errorMsg(对齐 Java)。

const (
	transferCreatePath         = "/v1/fund_trade/mch-transfer/transfer-bills"
	transferQueryByBillNoPath  = "/v1/fund_trade/mch-transfer/transfer-bills/transfer-bill-no/%s"
	transferQueryByOutBillPath = "/v1/fund_trade/mch-transfer/transfer-bills/out-bill-no/%s"
	profitSharingCreatePath    = "/v1/trade/profitsharing/orders"
	profitSharingQueryPath     = "/v1/trade/profitsharing/orders/%s"
)

// Transfer：发起商家转账
func Transfer(ctx context.Context, req *dto.TransferReq) (*dto.TransferResp, error) {
	middleware.LoggerWithTrace(ctx).Info("douyin transfer", "outBillNo", req.OutBillNo, "amount", req.Amount)
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"appid":             req.Credential.DouyinAppID,
		"out_bill_no":       req.OutBillNo,
		"transfer_scene_id": req.Scene,
		// 转账金额: 抖音要求整数单位分, req.Amount 已为分
		"transfer_amount": int64(req.Amount),
		"transfer_remark": truncateRunes(req.Remark, 32),
	}
	if req.NotifyURL != "" {
		body["notify_url"] = req.NotifyURL
	}
	// 收款人: openid 与手机号二选一(手机号走敏感字段加密)
	if req.PhoneNumber != "" {
		body["phone_number"] = req.PhoneNumber
	} else {
		body["openid"] = req.Openid
	}
	if req.Perception != "" {
		body["user_recv_perception"] = truncateRunes(req.Perception, 64)
	}
	// 转账场景报备信息(按场景要求的 info_type 填写)
	if len(req.ReportInfos) > 0 {
		infos := make([]map[string]any, 0, len(req.ReportInfos))
		for _, info := range req.ReportInfos {
			infos = append(infos, map[string]any{
				"info_type":    info.InfoType,
				"info_content": info.InfoContent,
			})
		}
		body["transfer_scene_report_infos"] = infos
	}

	// 敏感字段(收款人姓名/手机号): 平台证书 RSA 加密, 并携带证书序列号头
	var headers map[string]string
	if req.UserName != "" {
		encrypted, serial, err := encryptWithPlatformCert(ctx, client, req.UserName)
		if err != nil {
			return nil, douyin.NewSDKError("channel.error.douyinTransferEncryptFailed", err.Error())
		}
		body["user_name"] = encrypted
		headers = map[string]string{"Douyinpay-Serial": serial}
	}
	if req.PhoneNumber != "" {
		encrypted, serial, err := encryptWithPlatformCert(ctx, client, req.PhoneNumber)
		if err != nil {
			return nil, douyin.NewSDKError("channel.error.douyinTransferEncryptFailed", err.Error())
		}
		body["phone_number"] = encrypted
		if headers == nil {
			headers = map[string]string{"Douyinpay-Serial": serial}
		} else {
			headers["Douyinpay-Serial"] = serial
		}
	}

	raw, err := client.DoWithHeaders(ctx, http.MethodPost, transferCreatePath, body, headers)
	if err != nil {
		return nil, wrapAPIErr("channel.error.douyinTransferFailed", err)
	}
	var out struct {
		TransferBillNo string `json:"transfer_bill_no"`
		State          string `json:"state"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, douyin.NewSDKError("channel.error.douyinTransferFailed", err.Error())
	}
	return &dto.TransferResp{TransferBillNo: out.TransferBillNo, State: out.State}, nil
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
		path = fmt.Sprintf(transferQueryByBillNoPath, req.OutBillNo)
	} else {
		path = fmt.Sprintf(transferQueryByOutBillPath, req.TransferNo)
	}
	raw, err := client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, wrapAPIErr("channel.error.douyinTransferQueryFailed", err)
	}
	var out struct {
		TransferBillNo string `json:"transfer_bill_no"`
		State          string `json:"state"`
		FailReason     string `json:"fail_reason"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, douyin.NewSDKError("channel.error.douyinTransferQueryFailed", err.Error())
	}
	return &dto.TransferResp{
		TransferBillNo: out.TransferBillNo,
		State:          out.State,
		FailReason:     out.FailReason,
	}, nil
}

// Alloc：发起分账(splitFund, unfreeze_unsplit=true 自动解冻剩余)
func Alloc(ctx context.Context, req *dto.AllocReq) (*dto.AllocResp, error) {
	middleware.LoggerWithTrace(ctx).Info("douyin alloc", "outTradeNo", req.OutTradeNo, "tradeNo", req.TradeNo)
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	// 接收方列表: 个人类型姓名(PERSONAL_OPENID)为敏感字段, 平台证书 RSA 加密
	receivers := make([]map[string]any, 0, len(req.ReceiverInfoDtos))
	var headers map[string]string
	for _, r := range req.ReceiverInfoDtos {
		item := map[string]any{
			"type":    r.Type,
			"account": r.Account,
			"amount":  int64(r.Amount),
		}
		if r.Name != "" {
			encrypted, serial, err := encryptWithPlatformCert(ctx, client, r.Name)
			if err != nil {
				return &dto.AllocResp{ErrorCode: err.Error(), ErrorMsg: err.Error()}, nil
			}
			item["name"] = encrypted
			if headers == nil {
				headers = map[string]string{"Douyinpay-Serial": serial}
			} else {
				headers["Douyinpay-Serial"] = serial
			}
		}
		receivers = append(receivers, item)
	}
	body := map[string]any{
		"appid":            req.Credential.DouyinAppID,
		"mchid":            req.Credential.MchID,
		"transaction_id":   req.TradeNo,
		"out_order_no":     req.OutTradeNo,
		"unfreeze_unsplit": true,
		"receivers":        receivers,
	}
	if req.NotifyURL != "" {
		body["notify_url"] = req.NotifyURL
	}

	raw, err := client.DoWithHeaders(ctx, http.MethodPost, profitSharingCreatePath, body, headers)
	if err != nil {
		// 异常不抛, 包 resp 返回 errorCode/errorMsg(对齐 Java 两个字段均填 message)
		return &dto.AllocResp{ErrorCode: err.Error(), ErrorMsg: err.Error()}, nil
	}
	var out struct {
		OrderID string `json:"order_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return &dto.AllocResp{ErrorCode: err.Error(), ErrorMsg: err.Error()}, nil
	}
	return &dto.AllocResp{OrderId: out.OrderID}, nil
}

// AllocSync：分账同步查询(querySplitFund)
func AllocSync(ctx context.Context, req *dto.AllocReq) (*dto.AllocResp, error) {
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf(profitSharingQueryPath, url.PathEscape(req.OutTradeNo)) +
		"?mchid=" + url.QueryEscape(req.Credential.MchID) +
		"&transaction_id=" + url.QueryEscape(req.TradeNo)
	raw, err := client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		// 异常不抛, 包 resp 返回 errorCode/errorMsg(对齐 Java 两个字段均填 message)
		return &dto.AllocResp{ErrorCode: err.Error(), ErrorMsg: err.Error()}, nil
	}
	var out struct {
		OrderID         string `json:"order_id"`
		State           string `json:"state"`
		SplitFinishTime string `json:"split_finish_time"`
		Receivers       []struct {
			Account    string `json:"account"`
			Amount     *int64 `json:"amount"`
			Result     string `json:"result"`
			FailReason string `json:"fail_reason"`
			FinishTime string `json:"finish_time"`
		} `json:"receivers"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return &dto.AllocResp{ErrorCode: err.Error(), ErrorMsg: err.Error()}, nil
	}
	resp := &dto.AllocResp{
		OrderId: out.OrderID,
		Status:  out.State,
	}
	// 映射逐明细结果
	for _, r := range out.Receivers {
		rr := dto.ReceiverSplitResult{
			Account:     r.Account,
			SplitStatus: r.Result,
			FailReason:  r.FailReason,
			FinishTime:  r.FinishTime,
		}
		if r.Amount != nil {
			rr.Amount = ptrInt64(*r.Amount)
		}
		resp.ReceiverSplitResultDtos = append(resp.ReceiverSplitResultDtos, rr)
	}
	return resp, nil
}

// encryptWithPlatformCert：平台证书 RSA(PKCS#1 v1.5) 加密敏感字段, 返回密文与证书序列号(十六进制大写)
func encryptWithPlatformCert(ctx context.Context, client *openapi.Client, plain string) (string, string, error) {
	pub, serial, err := client.PlatformCert(ctx)
	if err != nil {
		return "", "", err
	}
	encrypted, err := openapi.EncryptPKCS1v15(pub, plain)
	if err != nil {
		return "", "", err
	}
	return encrypted, serial, nil
}
