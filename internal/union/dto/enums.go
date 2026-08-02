package dto

// PayMethod：云闪付通道支付方式
type PayMethod string

const (
	// PayMethodQRCode：主扫支付(申请二维码, C 扫 B)
	PayMethodQRCode PayMethod = "QRCODE"
	// PayMethodBarcode：被扫支付(付款码消费, B 扫 C)
	PayMethodBarcode PayMethod = "BARCODE"
	// PayMethodH5：H5/WAP 网关支付(前台跳转)
	PayMethodH5 PayMethod = "H5"
)

// PayBodyType：支付体类型
type PayBodyType string

const (
	// PayBodyTypeQRCode：二维码内容(主扫)
	PayBodyTypeQRCode PayBodyType = "QR_CODE"
	// PayBodyTypeLink：跳转链接(H5)
	PayBodyTypeLink PayBodyType = "LINK"
)
