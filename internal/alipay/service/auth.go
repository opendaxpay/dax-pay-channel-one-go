package service

import (
	"context"

	"daxpay.open/dax-pay-channel-one-go/internal/alipay/dto"
)

// ExchangeAppAuthToken：授权码换 app_auth_token
func ExchangeAppAuthToken(ctx context.Context, req *dto.AppAuthTokenReq) (*dto.AppAuthTokenResp, error) {
	biz := map[string]any{
		"grant_type": "authorization_code",
		"code":       req.AuthCode,
	}
	var out struct {
		gatewayBiz
		AppAuthToken    string `json:"app_auth_token"`
		AppRefreshToken string `json:"app_refresh_token"`
		AuthAppID       string `json:"auth_app_id"`
		UserID          string `json:"user_id"`
		ExpiresIn       string `json:"expires_in"`
		ReExpiresIn     string `json:"re_expires_in"`
	}
	if err := executeJSON(ctx, req.Credential, "alipay.open.auth.token.app", biz, "", &out); err != nil {
		return nil, wrapOr(err, "channel.error.alipayAppAuthTokenFailed")
	}
	// 对标 Boot：原样回传字段（含失败 code），不因业务码非 10000 直接抛错
	return &dto.AppAuthTokenResp{
		Code:            out.Code,
		SubCode:         out.SubCode,
		SubMsg:          out.SubMsg,
		AppAuthToken:    out.AppAuthToken,
		AppRefreshToken: out.AppRefreshToken,
		AuthAppID:       out.AuthAppID,
		UserID:          out.UserID,
		ExpiresIn:       out.ExpiresIn,
		ReExpiresIn:     out.ReExpiresIn,
	}, nil
}
