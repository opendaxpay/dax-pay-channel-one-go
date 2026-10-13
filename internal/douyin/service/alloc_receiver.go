package service

import (
	"context"
	"net/http"
	"strings"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/douyin/openapi"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// 抖音分账接收方服务(对标 Java DouyinAllocReceiverService)
//
// 接收方通道侧注册/删除(addSplitReceiver / deleteSplitReceiver), 走 /v1/trade/profitsharing/receivers/*。
// 绑定/解绑均为同步调用, 失败不抛异常, 包 resp 返回 errorCode/errorMsg 由主应用决策;
// 解绑遇"分账接收方不存在"视为幂等成功(对齐支付宝 USER_NOT_EXIST 容错)。

const (
	profitSharingReceiverAddPath    = "/v1/trade/profitsharing/receivers/add"
	profitSharingReceiverDeletePath = "/v1/trade/profitsharing/receivers/delete"
)

// BindReceiver：绑定分账接收方(addSplitReceiver)
//
// 接收方名称为敏感字段(MERCHANT_ID 时为商户全称/PERSONAL_OPENID 时为个人姓名),
// 平台证书 RSA(PKCS#1 v1.5) 加密上送并携带 Douyinpay-Serial 头(Java 侧由 SDK 内部完成);
// 关系类型由平台小写转抖音原生大写。
func BindReceiver(ctx context.Context, req *dto.AllocReceiverReq) (*dto.AllocReceiverResp, error) {
	middleware.LoggerWithTrace(ctx).Info("douyin alloc receiver bind", "account", req.ReceiverAccount)
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"mchid":   req.Credential.MchID,
		"appid":   req.Credential.DouyinAppID,
		"type":    req.ReceiverType,
		"account": req.ReceiverAccount,
		// 平台小写关系类型转抖音原生大写
		"relation_type": strings.ToUpper(req.RelationType),
	}
	// 自定义分账关系名(relationType=CUSTOM 时必填), 为空不传
	if req.CustomRelation != "" {
		body["custom_relation"] = req.CustomRelation
	}
	// 敏感字段(接收方名称): 平台证书 RSA 加密, 并携带证书序列号头
	var headers map[string]string
	if req.ReceiverName != "" {
		encrypted, serial, err := encryptWithPlatformCert(ctx, client, req.ReceiverName)
		if err != nil {
			// 加密失败不抛, 包 resp(对齐 Java catch DouyinpayException 语义)
			return &dto.AllocReceiverResp{ErrorCode: err.Error(), ErrorMsg: err.Error()}, nil
		}
		body["name"] = encrypted
		headers = map[string]string{"Douyinpay-Serial": serial}
	}
	// 成功响应无业务数据, 不解析响应体
	if _, err := client.DoWithHeaders(ctx, http.MethodPost, profitSharingReceiverAddPath, body, headers); err != nil {
		middleware.LoggerWithTrace(ctx).Error("douyin alloc receiver bind failed", "account", req.ReceiverAccount, "err", err.Error())
		return allocReceiverErrorResp(err), nil
	}
	return &dto.AllocReceiverResp{}, nil
}

// UnbindReceiver：解绑分账接收方(deleteSplitReceiver)
//
// 解绑遇"分账接收方不存在"视为幂等成功(通道侧已无绑定关系, 返回空成功响应);
// 其余失败不抛, 包 resp 返回 errorCode/errorMsg。
func UnbindReceiver(ctx context.Context, req *dto.AllocReceiverReq) (*dto.AllocReceiverResp, error) {
	middleware.LoggerWithTrace(ctx).Info("douyin alloc receiver unbind", "account", req.ReceiverAccount)
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"mchid":   req.Credential.MchID,
		"appid":   req.Credential.DouyinAppID,
		"type":    req.ReceiverType,
		"account": req.ReceiverAccount,
	}
	if _, err := client.Do(ctx, http.MethodPost, profitSharingReceiverDeletePath, body); err != nil {
		// 分账接收方不存在视为幂等成功(通道侧已无绑定关系)
		if isReceiverNotExist(err) {
			middleware.LoggerWithTrace(ctx).Info("douyin alloc receiver unbind idempotent success (receiver not exist)", "account", req.ReceiverAccount)
			return &dto.AllocReceiverResp{}, nil
		}
		middleware.LoggerWithTrace(ctx).Error("douyin alloc receiver unbind failed", "account", req.ReceiverAccount, "err", err.Error())
		return allocReceiverErrorResp(err), nil
	}
	return &dto.AllocReceiverResp{}, nil
}

// isReceiverNotExist：是否为"分账接收方不存在"(解绑幂等容错, 对齐 Java isReceiverNotExist)
//
// 抖音对该场景无文档化错误码, 按异常消息组合匹配:
// 须同时含中文「接收方」与「不存在」, 或同时含英文 receiver 与 not exist / not_exist, 任一组合命中即认定;
// 组合双词匹配刻意避免误吞「商户不存在」等其他错误。
func isReceiverNotExist(err error) bool {
	if err == nil {
		return false
	}
	text := err.Error()
	cnHit := strings.Contains(text, "接收方") && strings.Contains(text, "不存在")
	lower := strings.ToLower(text)
	enHit := strings.Contains(lower, "receiver") &&
		(strings.Contains(lower, "not exist") || strings.Contains(lower, "not_exist"))
	return cnHit || enHit
}

// allocReceiverErrorResp：分账接收方绑定/解绑异常 → 包 resp(对齐 Java catch 后填充错误信息的语义)
//
// APIError(网关已结构化)取通道错误码 + "message: detail.issue" 简洁文案(与分账 [allocErrorResp] 同语义);
// 其他异常(签名/网络类)无结构化信息, 退回原始 message; 完整调试串仅留在日志, 不透传给主应用。
func allocReceiverErrorResp(err error) *dto.AllocReceiverResp {
	resp := &dto.AllocReceiverResp{}
	if apiErr, ok := err.(*openapi.APIError); ok {
		resp.ErrorCode = apiErr.Code
		resp.ErrorMsg = apiErr.Message
		if issue := detailIssue(apiErr.Body); issue != "" {
			resp.ErrorMsg = apiErr.Message + ": " + issue
		}
		if resp.ErrorMsg == "" {
			resp.ErrorMsg = apiErr.Error()
		}
		return resp
	}
	resp.ErrorCode = err.Error()
	resp.ErrorMsg = err.Error()
	return resp
}
