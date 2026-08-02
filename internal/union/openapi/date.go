package openapi

import "time"

// 东八区时区(北京时间)
var cstZone = time.FixedZone("CST", 8*3600)

// TxnTime：当前东八区紧凑时间(yyyyMMddHHmmss), 用于银联 txnTime
func TxnTime() string {
	return time.Now().In(cstZone).Format("20060102150405")
}
