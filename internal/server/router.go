package server

import (
	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/handler"
	"daxpay.open/dax-pay-channel-one-go/internal/middleware"
	"daxpay.open/dax-pay-channel-one-go/internal/transport"
)

const ServiceName = "daxpay-channel-one-go"

// NewRouter：注册中间件与路由
//
// 中间件顺序：Recovery → Otel → TraceIDHeader → Locale → TransportEncrypt
// TransportEncrypt 置于 Locale 之后（locale/trace 先就绪，解密失败可用本地化文案）。
func NewRouter(encryptor *transport.Encryptor) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Otel(ServiceName))
	r.Use(middleware.TraceIDHeader())
	r.Use(middleware.Locale())
	r.Use(middleware.TransportEncrypt(encryptor))

	r.GET("/actuator/health", handler.Health)
	r.GET("/internal/probe", handler.Probe)

	// 支付宝（与 Boot AlipayChannelClient 路径一致）
	alipay := r.Group("/channel/alipay")
	{
		alipay.POST("/pay", handler.AlipayPay)
		alipay.POST("/sync", handler.AlipaySync)
		alipay.POST("/close", handler.AlipayClose)
		alipay.POST("/refund", handler.AlipayRefund)
		alipay.POST("/refund-sync", handler.AlipayRefundSync)
		alipay.POST("/transfer", handler.AlipayTransfer)
		alipay.POST("/transfer-sync", handler.AlipayTransferSync)
		alipay.POST("/alloc", handler.AlipayAlloc)
		alipay.POST("/alloc-sync", handler.AlipayAllocSync)
		alipay.POST("/callback/parse-pay", handler.AlipayCallbackParsePay)
		alipay.POST("/callback/parse-refund", handler.AlipayCallbackParseRefund)
		alipay.POST("/callback/parse-transfer", handler.AlipayCallbackParseTransfer)
		alipay.POST("/auth/app-token", handler.AlipayAppAuthToken)
	}

	// 微信直连
	wechat := r.Group("/channel/wechat")
	{
		wechat.POST("/pay", handler.WechatPay)
		wechat.POST("/sync", handler.WechatSync)
		wechat.POST("/close", handler.WechatClose)
		wechat.POST("/refund", handler.WechatRefund)
		wechat.POST("/refund-sync", handler.WechatRefundSync)
		wechat.POST("/transfer", handler.WechatTransfer)
		wechat.POST("/transfer-sync", handler.WechatTransferSync)
		wechat.POST("/alloc", handler.WechatAlloc)
		wechat.POST("/alloc-sync", handler.WechatAllocSync)
		wechat.POST("/callback/parse-pay", handler.WechatCallbackParsePay)
		wechat.POST("/callback/parse-refund", handler.WechatCallbackParseRefund)
		wechat.POST("/callback/parse-transfer", handler.WechatCallbackParseTransfer)
	}

	// 微信服务商（ISV，无 callback）
	wechatIsv := r.Group("/channel/wechat/isv")
	{
		wechatIsv.POST("/pay", handler.WechatIsvPay)
		wechatIsv.POST("/sync", handler.WechatIsvSync)
		wechatIsv.POST("/close", handler.WechatIsvClose)
		wechatIsv.POST("/refund", handler.WechatIsvRefund)
		wechatIsv.POST("/refund-sync", handler.WechatIsvRefundSync)
	}

	// 银联商务 UMS
	ums := r.Group("/channel/ums")
	{
		ums.POST("/pay", handler.UmsPay)
		ums.POST("/sync", handler.UmsSync)
		ums.POST("/close", handler.UmsClose)
		ums.POST("/refund", handler.UmsRefund)
		ums.POST("/refund-sync", handler.UmsRefundSync)
		ums.POST("/callback/parse-pay", handler.UmsCallbackParsePay)
		ums.POST("/callback/parse-refund", handler.UmsCallbackParseRefund)
	}

	// 云闪付(直连银联 ACP)
	union := r.Group("/channel/union")
	{
		union.POST("/pay", handler.UnionPay)
		union.POST("/sync", handler.UnionSync)
		union.POST("/close", handler.UnionClose)
		union.POST("/refund", handler.UnionRefund)
		union.POST("/refund-sync", handler.UnionRefundSync)
		union.POST("/callback/parse-pay", handler.UnionCallbackParsePay)
		union.POST("/callback/parse-refund", handler.UnionCallbackParseRefund)
	}

	// 抖音
	douyin := r.Group("/channel/douyin")
	{
		douyin.POST("/pay", handler.DouyinPay)
		douyin.POST("/sync", handler.DouyinSync)
		douyin.POST("/close", handler.DouyinClose)
		douyin.POST("/refund", handler.DouyinRefund)
		douyin.POST("/refund-sync", handler.DouyinRefundSync)
		douyin.POST("/transfer", handler.DouyinTransfer)
		douyin.POST("/transfer-sync", handler.DouyinTransferSync)
		douyin.POST("/alloc", handler.DouyinAlloc)
		douyin.POST("/alloc-sync", handler.DouyinAllocSync)
		douyin.POST("/callback/parse-pay", handler.DouyinCallbackParsePay)
		douyin.POST("/callback/parse-refund", handler.DouyinCallbackParseRefund)
		douyin.POST("/callback/parse-transfer", handler.DouyinCallbackParseTransfer)
		douyin.POST("/callback/parse-alloc", handler.DouyinCallbackParseAlloc)
	}

	return r
}
