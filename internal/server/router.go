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

	// 支付宝占位路径（与 Boot AlipayChannelClient / AlipayPayController 一致）
	alipay := r.Group("/channel/alipay")
	{
		alipay.POST("/pay", handler.AlipayStub)
		alipay.POST("/sync", handler.AlipayStub)
		alipay.POST("/close", handler.AlipayStub)
		alipay.POST("/refund", handler.AlipayStub)
		alipay.POST("/refund-sync", handler.AlipayStub)
		alipay.POST("/callback/parse-pay", handler.AlipayStub)
		alipay.POST("/callback/parse-refund", handler.AlipayStub)
		alipay.POST("/auth/app-token", handler.AlipayStub)
	}

	return r
}
