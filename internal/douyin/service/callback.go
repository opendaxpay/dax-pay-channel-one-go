package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"daxpay.open/dax-pay-channel-one-go/internal/douyin/dto"
	"daxpay.open/dax-pay-channel-one-go/internal/douyin/openapi"
	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
)

// ParsePayCallback：支付回调验签解密
func ParsePayCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	return parseCallback(ctx, req, "PAY")
}

// ParseRefundCallback：退款回调验签解密
func ParseRefundCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.CallbackParseResp {
	return parseCallback(ctx, req, "REFUND")
}

// ParseTransferCallback：转账回调验签解密(商家转账异步通知)
//
// 通知体仅含 order_id(通道转账单号), 不含商户单号 out_bill_no。
func ParseTransferCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.TransferCallbackParseResp {
	resp := &dto.TransferCallbackParseResp{}
	if req.Credential == nil {
		return resp
	}
	client, err := newClient(req.Credential)
	if err != nil {
		return resp
	}
	if err := client.VerifyNotify(ctx, req.Timestamp, req.Nonce, req.Body, req.Signature, req.Serial); err != nil {
		return resp
	}
	plain, ok := decryptResource(client, req.Body)
	if !ok {
		return resp
	}
	var data struct {
		OrderID     string `json:"order_id"`
		Status      string `json:"status"`
		StatusDesc  string `json:"status_desc"`
		SuccessTime string `json:"success_time"`
	}
	if json.Unmarshal(plain, &data) != nil {
		return resp
	}
	resp.Verified = true
	resp.TransferBillNo = data.OrderID
	resp.TransferState = data.Status
	resp.TransferStatusDesc = data.StatusDesc
	resp.SuccessTime = data.SuccessTime
	return resp
}

// ParseAllocCallback：分账回调验签解密
//
// 通知体与分账查询响应同构(复用 receivers 逐明细解析)。
func ParseAllocCallback(ctx context.Context, req *dto.CallbackParseReq) *dto.AllocCallbackParseResp {
	resp := &dto.AllocCallbackParseResp{}
	if req.Credential == nil {
		return resp
	}
	client, err := newClient(req.Credential)
	if err != nil {
		return resp
	}
	if err := client.VerifyNotify(ctx, req.Timestamp, req.Nonce, req.Body, req.Signature, req.Serial); err != nil {
		return resp
	}
	plain, ok := decryptResource(client, req.Body)
	if !ok {
		return resp
	}
	var data struct {
		OrderID         string `json:"order_id"`
		State           string `json:"state"`
		SplitFinishTime string `json:"split_finish_time"`
		Receivers       []struct {
			Account    string `json:"account"`
			Result     string `json:"result"`
			FailReason string `json:"fail_reason"`
			FinishTime string `json:"finish_time"`
		} `json:"receivers"`
	}
	if json.Unmarshal(plain, &data) != nil {
		return resp
	}
	resp.Verified = true
	resp.OrderId = data.OrderID
	resp.State = data.State
	resp.SplitFinishTime = data.SplitFinishTime
	// 映射逐明细结果
	for _, r := range data.Receivers {
		resp.ReceiverResults = append(resp.ReceiverResults, dto.AllocCallbackReceiverResult{
			Account:     r.Account,
			SplitStatus: r.Result,
			FailReason:  r.FailReason,
			FinishTime:  r.FinishTime,
		})
	}
	return resp
}

// decryptResource：验签通过后解密回调 resource, 失败返回 false
func decryptResource(client *openapi.Client, body string) ([]byte, bool) {
	var envelope struct {
		Resource struct {
			Ciphertext     string `json:"ciphertext"`
			AssociatedData string `json:"associated_data"`
			Nonce          string `json:"nonce"`
		} `json:"resource"`
	}
	if json.Unmarshal([]byte(body), &envelope) != nil {
		return nil, false
	}
	plain, err := client.DecryptResource(envelope.Resource.AssociatedData, envelope.Resource.Nonce, envelope.Resource.Ciphertext)
	if err != nil {
		return nil, false
	}
	return plain, true
}

func parseCallback(ctx context.Context, req *dto.CallbackParseReq, tradeType string) *dto.CallbackParseResp {
	resp := &dto.CallbackParseResp{TradeType: tradeType, Verified: false}
	if req.Credential == nil {
		return resp
	}
	client, err := newClient(req.Credential)
	if err != nil {
		return resp
	}
	if err := client.VerifyNotify(ctx, req.Timestamp, req.Nonce, req.Body, req.Signature, req.Serial); err != nil {
		return resp
	}
	var envelope struct {
		Resource struct {
			Ciphertext     string `json:"ciphertext"`
			AssociatedData string `json:"associated_data"`
			Nonce          string `json:"nonce"`
		} `json:"resource"`
	}
	if json.Unmarshal([]byte(req.Body), &envelope) != nil {
		return resp
	}
	plain, err := client.DecryptResource(envelope.Resource.AssociatedData, envelope.Resource.Nonce, envelope.Resource.Ciphertext)
	if err != nil {
		return resp
	}
	var data map[string]any
	if json.Unmarshal(plain, &data) != nil {
		return resp
	}
	resp.Verified = true
	if tradeType == "PAY" {
		resp.OutTradeNo = anyString(data["out_trade_no"])
		resp.TransactionID = anyString(data["transaction_id"])
		resp.TradeState = anyString(data["trade_state"])
		resp.SuccessTime = anyString(data["success_time"])
		if amount, ok := data["amount"].(map[string]any); ok {
			resp.Amount = anyInt64Ptr(amount["total"])
		}
		if payer, ok := data["payer"].(map[string]any); ok {
			resp.OpenID = anyString(payer["openid"])
		}
	} else {
		resp.OutTradeNo = anyString(data["out_trade_no"])
		resp.OutRefundNo = anyString(data["out_refund_no"])
		resp.RefundID = anyString(data["refund_id"])
		resp.RefundStatus = anyString(data["refund_status"])
		resp.SuccessTime = anyString(data["success_time"])
		if amount, ok := data["amount"].(map[string]any); ok {
			resp.Amount = anyInt64Ptr(amount["refund"])
		}
	}
	return resp
}

func anyString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func anyInt64Ptr(v any) *jsonx.Int64String {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case float64:
		return ptrInt64(int64(t))
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil
		}
		return ptrInt64(n)
	default:
		return nil
	}
}
