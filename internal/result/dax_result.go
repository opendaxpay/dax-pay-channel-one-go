package result

// 统一响应封装（对标 Boot DaxResult）
//
// 子应用所有接口统一返回本对象。主应用只凭 body.code 判断成败，故 HTTP 状态码始终为 200。
// traceId 不写入响应体，由中间件通过 x-trace-id 响应头返回。
type DaxResult[T any] struct {
	// 业务状态码（0 成功，非 0 为错误码）
	Code int `json:"code"`
	// 提示信息（已按当前 locale 本地化）
	Msg string `json:"msg"`
	// 业务数据；失败时常省略
	Data *T `json:"data,omitempty"`
}

// 非泛型失败响应（data 省略）
type FailResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// 构建成功响应（msg 由调用方填入本地化文案）
func Ok[T any](msg string, data T) DaxResult[T] {
	return DaxResult[T]{
		Code: 0,
		Msg:  msg,
		Data: &data,
	}
}

// 构建失败响应
func Fail(code int, msg string) FailResult {
	return FailResult{Code: code, Msg: msg}
}
