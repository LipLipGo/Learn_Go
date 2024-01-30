package web

// 定义统一的响应格式

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"` // 如果时候数字的话要使用 float64，因为数字转 any 的话，默认是 float64
}
