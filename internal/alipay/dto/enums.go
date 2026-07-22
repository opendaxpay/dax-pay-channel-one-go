package dto

// PayMethod：支付方式（枚举 name）
type PayMethod string

const (
	MethodWAP     PayMethod = "WAP"
	MethodAPP     PayMethod = "APP"
	MethodPC      PayMethod = "PC"
	MethodQR      PayMethod = "QR"
	MethodBARCODE PayMethod = "BARCODE"
	MethodJSAPI   PayMethod = "JSAPI"
)

// PayBodyType：支付体类型
type PayBodyType string

const (
	BodyLINK       PayBodyType = "LINK"
	BodyQRCode     PayBodyType = "QR_CODE"
	BodyOrderStr   PayBodyType = "ORDER_STR"
	BodyIdentifier PayBodyType = "IDENTIFIER"
)
