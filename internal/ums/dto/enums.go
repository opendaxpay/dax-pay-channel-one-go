package dto

// PayMethod：银联商务支付方式
type PayMethod string

const (
	PayMethodQRCode        PayMethod = "QRCODE"
	PayMethodAlipayH5      PayMethod = "ALIPAY_H5"
	PayMethodWechatH5      PayMethod = "WECHAT_H5"
	PayMethodWechatCashier PayMethod = "WECHAT_CASHIER"
	PayMethodUnionJSAPI    PayMethod = "UNION_JSAPI"
)

// PayBodyType：支付体类型
type PayBodyType string

const (
	PayBodyTypeQRCode PayBodyType = "QR_CODE"
	PayBodyTypeLink   PayBodyType = "LINK"
)

const (
	InstMidQR = "QRPAYDEFAULT"
	InstMidH5 = "H5DEFAULT"
)
