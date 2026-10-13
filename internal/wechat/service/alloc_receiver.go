package service

import (
	"context"
	"net/http"
	"strings"

	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/wechat/openapi"
)

// 微信分账接收方服务(直连 + 服务商 ISV, 对标 Java WechatDirect/WechatIsvAllocReceiverService)
//
// 接收方通道侧注册/删除(V3 /v3/profitsharing/receivers/add / delete), 直连与服务商同一路径,
// 服务商报文多 sub_mchid(+可选 sub_appid, 特约商户维度)。
// 绑定/解绑均为同步调用, 失败不抛异常, 包 resp 返回 errorCode/errorMsg 由主应用决策;
// 解绑遇"接收方不存在"视为幂等成功(错误码精确匹配 + 描述组合匹配, 对齐 Java)。
// 注: 微信 ISV 的分账执行链路(profitsharing/orders)未接, 本服务仅覆盖接收方注册(对齐 Java)。

const (
	profitSharingReceiverAddPath    = "/v3/profitsharing/receivers/add"
	profitSharingReceiverDeletePath = "/v3/profitsharing/receivers/delete"
)

// 微信"资源不存在"类错误码(解绑幂等容错)
const (
	errCodeNotFound          = "NOT_FOUND"
	errCodeResourceNotExists = "RESOURCE_NOT_EXISTS"
)

// DirectAllocReceiverBind：直连绑定分账接收方(商户应用 appid 维度)
func DirectAllocReceiverBind(ctx context.Context, req *dto.AllocReceiverReq) (*dto.AllocReceiverResp, error) {
	return allocReceiverOp(ctx, req, false, profitSharingReceiverAddPath)
}

// IsvAllocReceiverBind：服务商绑定分账接收方(sp_appid + sub_mchid 维度)
func IsvAllocReceiverBind(ctx context.Context, req *dto.AllocReceiverReq) (*dto.AllocReceiverResp, error) {
	return allocReceiverOp(ctx, req, true, profitSharingReceiverAddPath)
}

// DirectAllocReceiverUnbind：直连解绑分账接收方
func DirectAllocReceiverUnbind(ctx context.Context, req *dto.AllocReceiverReq) (*dto.AllocReceiverResp, error) {
	return allocReceiverOp(ctx, req, false, profitSharingReceiverDeletePath)
}

// IsvAllocReceiverUnbind：服务商解绑分账接收方
func IsvAllocReceiverUnbind(ctx context.Context, req *dto.AllocReceiverReq) (*dto.AllocReceiverResp, error) {
	return allocReceiverOp(ctx, req, true, profitSharingReceiverDeletePath)
}

// allocReceiverOp：绑定/解绑共用实现(直连与服务商仅身份字段差异, 对齐 Java 两服务同构 fillReceiver)
func allocReceiverOp(ctx context.Context, req *dto.AllocReceiverReq, isv bool, path string) (*dto.AllocReceiverResp, error) {
	client, err := newClient(req.Credential)
	if err != nil {
		return nil, err
	}
	body, err := allocReceiverBody(ctx, client, req, isv)
	if err != nil {
		return &dto.AllocReceiverResp{ErrorCode: err.Error(), ErrorMsg: err.Error()}, nil
	}
	if _, err := client.Do(ctx, http.MethodPost, path, body); err != nil {
		// 解绑遇"分账接收方不存在"视为幂等成功(通道侧已无绑定关系)
		if path == profitSharingReceiverDeletePath && isWechatReceiverNotExist(err) {
			middleware.LoggerWithTrace(ctx).Info("wechat alloc receiver unbind idempotent success (receiver not exist)",
				"account", req.ReceiverAccount, "isv", isv)
			return &dto.AllocReceiverResp{}, nil
		}
		middleware.LoggerWithTrace(ctx).Error("wechat alloc receiver op failed",
			"account", req.ReceiverAccount, "isv", isv, "path", path, "err", err.Error())
		return allocReceiverErrorResp(err), nil
	}
	return &dto.AllocReceiverResp{}, nil
}

// allocReceiverBody：构造接收方报文(直连与服务商身份字段 + 接收方公共字段)
func allocReceiverBody(ctx context.Context, client *openapi.Client, req *dto.AllocReceiverReq, isv bool) (map[string]any, error) {
	cred := req.Credential
	body := map[string]any{
		// 直连为商户应用 appid; 服务商为 sp_appid(凭证 wxAppId) + sub_mchid 维度
		"appid":   cred.WxAppId,
		"type":    req.ReceiverType,
		"account": req.ReceiverAccount,
	}
	if isv {
		body["sub_mchid"] = cred.SubMchId
		if cred.SubAppId != "" {
			body["sub_appid"] = cred.SubAppId
		}
	}
	// 接收方名称为敏感字段(个人姓名), 微信支付平台证书 RSA-OAEP 加密上送(对齐 WxJava 敏感字段处理)
	if req.ReceiverName != "" {
		pub, err := client.PlatformCert(ctx)
		if err != nil {
			return nil, err
		}
		encrypted, err := openapi.EncryptOAEP(pub, req.ReceiverName)
		if err != nil {
			return nil, err
		}
		body["name"] = encrypted
	}
	// 关系类型映射: service_provider 为服务商保留值 → CUSTOM + 自定义关系名默认"服务商";
	// 其余转微信 V3 原生大写直传
	if req.RelationType == "service_provider" {
		body["relation_type"] = "CUSTOM"
		custom := req.CustomRelation
		if custom == "" {
			custom = "服务商"
		}
		body["custom_relation"] = custom
	} else {
		body["relation_type"] = strings.ToUpper(req.RelationType)
		if req.CustomRelation != "" {
			body["custom_relation"] = req.CustomRelation
		}
	}
	return body, nil
}

// isWechatReceiverNotExist：是否为"分账接收方不存在"(解绑幂等容错, 对齐 Java isReceiverNotExist)
//
// 微信对该场景无文档化错误码, 采用错误码精确匹配 + 描述组合匹配:
// 描述须同时含「接收方/receiver」与「不存在/not exist」语义, 避免误吞「商户不存在」等其他错误。
func isWechatReceiverNotExist(err error) bool {
	apiErr, ok := err.(*openapi.APIError)
	if !ok {
		return false
	}
	if apiErr.Code == errCodeNotFound || apiErr.Code == errCodeResourceNotExists {
		return true
	}
	text := apiErr.Message
	cnHit := strings.Contains(text, "接收方") && strings.Contains(text, "不存在")
	lower := strings.ToLower(text)
	enHit := strings.Contains(lower, "receiver") && strings.Contains(lower, "not exist")
	return cnHit || enHit
}

// allocReceiverErrorResp：绑定/解绑异常 → 包 resp(对齐 Java errorCode=errCode, errorMsg=errCodeDes 兜底 message)
func allocReceiverErrorResp(err error) *dto.AllocReceiverResp {
	resp := &dto.AllocReceiverResp{}
	if apiErr, ok := err.(*openapi.APIError); ok {
		resp.ErrorCode = apiErr.Code
		resp.ErrorMsg = apiErr.Message
		if resp.ErrorMsg == "" {
			resp.ErrorMsg = apiErr.Error()
		}
		return resp
	}
	resp.ErrorCode = err.Error()
	resp.ErrorMsg = err.Error()
	return resp
}
