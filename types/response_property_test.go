package types

import (
	"encoding/json"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Validates: Requirements 10.1, 10.2, 10.3, 10.4**
// Property 1: API 响应格式一致性
// 验证 Status 字段只能是 "ok" 或 "error"
// 验证成功响应包含 Data 字段
// 验证错误响应包含 Msg 字段

func TestProperty_APIResponseFormatConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 1.1: Status 字段只能是 "ok" 或 "error"
	properties.Property("Status field must be either 'ok' or 'error'", prop.ForAll(
		func(isSuccess bool, data string, msg string) bool {
			var response ApiResponse
			if isSuccess {
				response = SuccessResponse(data)
			} else {
				response = ErrorResponse(msg)
			}

			// Status 必须是 "ok" 或 "error"
			return response.Status == "ok" || response.Status == "error"
		},
		gen.Bool(),
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property 1.2: 成功响应的 Status 必须是 "ok"
	properties.Property("Success response must have status 'ok'", prop.ForAll(
		func(data string) bool {
			response := SuccessResponse(data)
			return response.Status == "ok"
		},
		gen.AlphaString(),
	))

	// Property 1.3: 错误响应的 Status 必须是 "error"
	properties.Property("Error response must have status 'error'", prop.ForAll(
		func(msg string) bool {
			response := ErrorResponse(msg)
			return response.Status == "error"
		},
		gen.AlphaString(),
	))

	// Property 1.4: 成功响应必须包含 Data 字段（非 nil 或有值）
	properties.Property("Success response must contain Data field", prop.ForAll(
		func(data string) bool {
			response := SuccessResponse(data)
			// 成功响应的 Data 字段应该被设置
			// 检查响应可以被序列化为 JSON 且包含 data 字段
			jsonBytes, err := json.Marshal(response)
			if err != nil {
				return false
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				return false
			}

			// 验证 data 字段存在于 JSON 中
			_, hasData := parsed["data"]
			return hasData
		},
		gen.AlphaString(),
	))

	// Property 1.5: 错误响应必须包含 Msg 字段
	properties.Property("Error response must contain Msg field", prop.ForAll(
		func(msg string) bool {
			response := ErrorResponse(msg)
			// 错误响应的 Msg 字段应该被设置
			return response.Msg == msg
		},
		gen.AlphaString(),
	))

	// Property 1.6: 成功响应不应该有 Msg 字段（或为空）
	properties.Property("Success response should not have Msg field", prop.ForAll(
		func(data string) bool {
			response := SuccessResponse(data)
			return response.Msg == ""
		},
		gen.AlphaString(),
	))

	// Property 1.7: 错误响应的 Data 字段应该为 nil（除非使用 ErrorResponseWithData）
	properties.Property("Error response should not have Data field by default", prop.ForAll(
		func(msg string) bool {
			response := ErrorResponse(msg)
			return response.Data == nil
		},
		gen.AlphaString(),
	))

	// Property 1.8: ErrorResponseWithData 必须同时包含 Msg 和 Data
	properties.Property("ErrorResponseWithData must contain both Msg and Data", prop.ForAll(
		func(msg string, data string) bool {
			response := ErrorResponseWithData(msg, data)
			// 验证 status, msg 和 data 字段
			if response.Status != "error" {
				return false
			}
			if response.Msg != msg {
				return false
			}
			// 验证 Data 字段被设置（通过 JSON 序列化检查）
			jsonBytes, err := json.Marshal(response)
			if err != nil {
				return false
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				return false
			}
			_, hasData := parsed["data"]
			return hasData
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property 1.9: 所有响应都可以被序列化为有效的 JSON
	properties.Property("All responses can be serialized to valid JSON", prop.ForAll(
		func(isSuccess bool, data string, msg string) bool {
			var response ApiResponse
			if isSuccess {
				response = SuccessResponse(data)
			} else {
				response = ErrorResponse(msg)
			}

			// 尝试序列化为 JSON
			jsonBytes, err := json.Marshal(response)
			if err != nil {
				return false
			}

			// 尝试反序列化回来
			var parsed ApiResponse
			err = json.Unmarshal(jsonBytes, &parsed)
			return err == nil
		},
		gen.Bool(),
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property 1.10: Status 字段永远不为空
	properties.Property("Status field is never empty", prop.ForAll(
		func(isSuccess bool, data string, msg string) bool {
			var response ApiResponse
			if isSuccess {
				response = SuccessResponse(data)
			} else {
				response = ErrorResponse(msg)
			}

			return response.Status != ""
		},
		gen.Bool(),
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// 测试响应格式的 JSON 序列化一致性
func TestProperty_JSONSerializationConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: JSON 序列化后的响应必须包含正确的字段
	properties.Property("JSON serialized response contains correct fields", prop.ForAll(
		func(isSuccess bool, data string, msg string) bool {
			var response ApiResponse
			if isSuccess {
				response = SuccessResponse(data)
			} else {
				response = ErrorResponse(msg)
			}

			jsonBytes, err := json.Marshal(response)
			if err != nil {
				return false
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				return false
			}

			// 验证 status 字段存在
			status, hasStatus := parsed["status"]
			if !hasStatus {
				return false
			}

			// 验证 status 值正确
			statusStr, ok := status.(string)
			if !ok {
				return false
			}

			if isSuccess {
				// 成功响应必须有 status="ok" 和 data 字段
				_, hasData := parsed["data"]
				return statusStr == "ok" && hasData
			} else {
				// 错误响应必须有 status="error"
				// msg 字段可能为空字符串时不会出现在 JSON 中（omitempty）
				// 但 Msg 字段在结构体中应该被设置
				return statusStr == "error" && response.Msg == msg
			}
		},
		gen.Bool(),
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// 测试响应类型的不变性
func TestProperty_ResponseImmutability(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	// Property: 创建响应后，Status 字段不应该改变
	properties.Property("Response Status field remains consistent", prop.ForAll(
		func(data string) bool {
			response := SuccessResponse(data)
			originalStatus := response.Status

			// 序列化和反序列化
			jsonBytes, _ := json.Marshal(response)
			var parsed ApiResponse
			json.Unmarshal(jsonBytes, &parsed)

			return parsed.Status == originalStatus && parsed.Status == "ok"
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// 测试不同数据类型的响应
func TestProperty_ResponseWithDifferentDataTypes(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: 字符串数据类型
	properties.Property("Success response with string data", prop.ForAll(
		func(data string) bool {
			response := SuccessResponse(data)
			return response.Status == "ok" && response.Data == data
		},
		gen.AlphaString(),
	))

	// Property: 整数数据类型
	properties.Property("Success response with integer data", prop.ForAll(
		func(data int) bool {
			response := SuccessResponse(data)
			return response.Status == "ok" && response.Data == data
		},
		gen.Int(),
	))

	// Property: 布尔数据类型
	properties.Property("Success response with boolean data", prop.ForAll(
		func(data bool) bool {
			response := SuccessResponse(data)
			return response.Status == "ok" && response.Data == data
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}
