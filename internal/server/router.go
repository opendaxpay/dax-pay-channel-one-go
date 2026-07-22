package server

import (
	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/handler"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
)

const ServiceName = "daxpay-channel-one-go"

// NewRouter：注册中间件与路由
//
// 中间件顺序：Recovery → Otel → TraceIDHeader → Locale
func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Otel(ServiceName))
	r.Use(middleware.TraceIDHeader())
	r.Use(middleware.Locale())

	r.GET("/actuator/health", handler.Health)
	r.GET("/internal/probe", handler.Probe)

	// 支付宝 OpenAPI 实装（与 Boot AlipayChannelClient 路径一致）
	alipay := r.Group("/channel/alipay")
	{
		alipay.POST("/pay", handler.AlipayPay)
		alipay.POST("/sync", handler.AlipaySync)
		alipay.POST("/close", handler.AlipayClose)
		alipay.POST("/refund", handler.AlipayRefund)
		alipay.POST("/refund-sync", handler.AlipayRefundSync)
		alipay.POST("/callback/parse-pay", handler.AlipayCallbackParsePay)
		alipay.POST("/callback/parse-refund", handler.AlipayCallbackParseRefund)
		alipay.POST("/auth/app-token", handler.AlipayAppAuthToken)
	}

	return r
}
