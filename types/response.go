package types

// ApiResponse 统一的 API 响应格式
// 所有 API 响应都遵循此格式，与现有前端期望的格式保持一致
type ApiResponse struct {
	Status string      `json:"status"`          // "ok" 或 "error"
	Msg    string      `json:"msg,omitempty"`   // 错误消息（仅在 Status 为 "error" 时使用）
	Data   interface{} `json:"data,omitempty"`  // 响应数据（仅在 Status 为 "ok" 时使用）
	Data2  interface{} `json:"data2,omitempty"` // 额外的响应数据（可选）
}

// SuccessResponse 创建成功响应
// 参数:
//   - data: 要返回的数据，可以是任何可序列化为 JSON 的类型
// 返回:
//   - ApiResponse: Status 为 "ok"，Data 字段包含传入的数据
func SuccessResponse(data interface{}) ApiResponse {
	return ApiResponse{
		Status: "ok",
		Data:   data,
	}
}

// ErrorResponse 创建错误响应
// 参数:
//   - msg: 错误消息字符串
// 返回:
//   - ApiResponse: Status 为 "error"，Msg 字段包含错误消息
func ErrorResponse(msg string) ApiResponse {
	return ApiResponse{
		Status: "error",
		Msg:    msg,
	}
}

// ErrorResponseWithData 创建带数据的错误响应
// 参数:
//   - msg: 错误消息字符串
//   - data: 额外的错误数据（例如验证错误详情）
// 返回:
//   - ApiResponse: Status 为 "error"，Msg 字段包含错误消息，Data 字段包含额外数据
func ErrorResponseWithData(msg string, data interface{}) ApiResponse {
	return ApiResponse{
		Status: "error",
		Msg:    msg,
		Data:   data,
	}
}
