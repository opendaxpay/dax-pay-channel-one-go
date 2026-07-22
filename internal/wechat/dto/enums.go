package dto

// PayMethod：微信支付方式（对标 WechatPayMethod）
type PayMethod string

const (
	PayMethodNative   PayMethod = "NATIVE"   // 扫码
	PayMethodJSAPI    PayMethod = "JSAPI"    // 公众号
	PayMethodMini     PayMethod = "MINI"     // 小程序（同 JSAPI 路径）
	PayMethodApp      PayMethod = "APP"      // APP
	PayMethodH5       PayMethod = "H5"       // H5
	PayMethodMicropay PayMethod = "MICROPAY" // 付款码（V3 codepay）
)

// PayBodyType：支付内容类型（对标 WechatPayBodyType）
type PayBodyType string

const (
	PayBodyLink        PayBodyType = "LINK"          // H5 跳转链接
	PayBodyQRCode      PayBodyType = "QR_CODE"       // NATIVE 二维码内容
	PayBodyJSAPI       PayBodyType = "JSAPI"         // JSAPI/小程序调起参数 JSON
	PayBodyAppOrderStr PayBodyType = "APP_ORDER_STR" // APP 调起参数 JSON
	PayBodyIdentifier  PayBodyType = "IDENTIFIER"    // 通用标识码（兜底）
)
