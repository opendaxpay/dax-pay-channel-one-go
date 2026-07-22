package dto

// PayMethod：抖音支付方式
type PayMethod string

const (
	PayMethodQR    PayMethod = "QR"
	PayMethodJSAPI PayMethod = "JSAPI"
	PayMethodH5    PayMethod = "H5"
	PayMethodAPP   PayMethod = "APP"
)

// PayBodyType：支付体类型
type PayBodyType string

const (
	PayBodyTypeQRCode     PayBodyType = "QR_CODE"
	PayBodyTypeLink       PayBodyType = "LINK"
	PayBodyTypeJSAPI      PayBodyType = "JSAPI"
	PayBodyTypeIdentifier PayBodyType = "IDENTIFIER"
)
