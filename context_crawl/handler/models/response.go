package models

// ============= 接口标准响应格式 ===================
type Response struct {
	Code int                    `json:"code"` // 0 成功 -1 失败
	Msg  string                 `json:"msg"`  // 成功/失败报错信息
	Data map[string]interface{} `json:"data"` // 任意map数据
}
