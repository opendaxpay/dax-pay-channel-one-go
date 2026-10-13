package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

// 支付宝分账接收方绑定/解绑服务(alipay.trade.royalty.relation.bind / unbind)
//
// 接收方通道侧注册/删除, 状态回写与失败处理由主应用决策。
// 业务失败码通过网关公共 code 判断, 失败时透传 subCode/subMsg 由主应用决策;
// 仅网络/解析类异常抛 SDK 调用失败(对齐 transfer.go 的错误处理口径)。
// 直连与服务商共用: 凭证含 appAuthToken 时由网关客户端自动以子商户身份代调用。

// subCodeUserNotExist：支付宝分账方不存在错误码(解绑幂等容错)
const subCodeUserNotExist = "USER_NOT_EXIST"

// BindReceiver：绑定分账接收方(alipay.trade.royalty.relation.bind)
func BindReceiver(ctx context.Context, req *dto.AllocReceiverReq) (*dto.AllocReceiverResp, error) {
	middleware.LoggerWithTrace(ctx).Info("alipay alloc receiver bind", "outRequestNo", req.OutRequestNo)
	// receiver_list 单元素: 绑定时名称非空则携带(对齐 Java StrUtil.isNotBlank 分支)
	biz := map[string]any{
		"out_request_no": req.OutRequestNo,
		"receiver_list":  []map[string]any{buildReceiverItem(req, true)},
	}
	var out struct {
		gatewayBiz
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.royalty.relation.bind", biz, "", &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayAllocFailed")
	}
	// 绑定无论成败均透传 subCode/subMsg(成功时为空), 由主应用按 errorCode 是否为空判定结果
	resp := &dto.AllocReceiverResp{ErrorCode: out.SubCode, ErrorMsg: out.SubMsg}
	if !out.success() {
		middleware.LoggerWithTrace(ctx).Warn("alipay alloc receiver bind biz failed",
			"outRequestNo", req.OutRequestNo, "code", out.Code, "subCode", out.SubCode, "subMsg", out.SubMsg)
	}
	return resp, nil
}

// UnbindReceiver：解绑分账接收方(alipay.trade.royalty.relation.unbind)
func UnbindReceiver(ctx context.Context, req *dto.AllocReceiverReq) (*dto.AllocReceiverResp, error) {
	middleware.LoggerWithTrace(ctx).Info("alipay alloc receiver unbind", "outRequestNo", req.OutRequestNo)
	// 解绑 receiver 仅含类型+账号, 不携带名称(对齐 Java unbind 不 set name)
	biz := map[string]any{
		"out_request_no": req.OutRequestNo,
		"receiver_list":  []map[string]any{buildReceiverItem(req, false)},
	}
	var out struct {
		gatewayBiz
	}
	if err := executeJSON(ctx, req.Credential, "alipay.trade.royalty.relation.unbind", biz, "", &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayAllocFailed")
	}
	resp := &dto.AllocReceiverResp{}
	if !out.success() {
		// 幂等容错: 分账方不存在(USER_NOT_EXIST)视为解绑成功——通道侧已不存在绑定关系,
		// 返回空 errorCode/errorMsg, 与解绑成功的响应形态一致, 避免主应用重试死循环
		if out.SubCode == subCodeUserNotExist {
			middleware.LoggerWithTrace(ctx).Info("alipay alloc receiver unbind idempotent success (receiver not exist)",
				"outRequestNo", req.OutRequestNo)
			return resp, nil
		}
		resp.ErrorCode = out.SubCode
		resp.ErrorMsg = out.SubMsg
		middleware.LoggerWithTrace(ctx).Warn("alipay alloc receiver unbind biz failed",
			"outRequestNo", req.OutRequestNo, "code", out.Code, "subCode", out.SubCode, "subMsg", out.SubMsg)
	}
	return resp, nil
}

// buildReceiverItem：构造 receiver_list 单元素
//
// 绑定时名称非空则携带(withName=true), 解绑仅类型+账号
func buildReceiverItem(req *dto.AllocReceiverReq, withName bool) map[string]any {
	item := map[string]any{
		// 平台大写枚举映射支付宝原生小写类型(userId/loginName)
		"type":    toAlipayReceiverType(req.ReceiverType),
		"account": req.ReceiverAccount,
	}
	if withName && req.ReceiverName != "" {
		item["name"] = req.ReceiverName
	}
	return item
}

// toAlipayReceiverType：平台接收方类型映射支付宝原生小写类型
//
// USER_ID → userId / LOGIN_NAME → loginName, 其余原样透传
func toAlipayReceiverType(receiverType string) string {
	switch receiverType {
	case "USER_ID":
		return "userId"
	case "LOGIN_NAME":
		return "loginName"
	default:
		return receiverType
	}
}
