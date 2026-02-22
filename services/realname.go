// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/oauth-server/oauth-server/models"
)

// RealNameVerifyRequest 实名认证请求
type RealNameVerifyRequest struct {
	Name   string `json:"name"`
	IDCard string `json:"idcard"`
}

// RealNameVerifyResponse 实名认证响应
type RealNameVerifyResponse struct {
	Success         bool                   `json:"success"`
	Message         string                 `json:"message"`
	OrderNo         string                 `json:"order_no"`
	AliyunAPICalled bool                   `json:"aliyun_api_called"`
	APIResponse     map[string]interface{} `json:"api_response,omitempty"`
}

// AliyunAPIResponse 阿里云API响应结构体
type AliyunAPIResponse struct {
	Code    int                    `json:"code"`
	Msg     string                 `json:"msg"`
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
}

// validateIDCardChecksum 使用 ISO 7064:1983, MOD 11-2 算法验证身份证号码校验位
func validateIDCardChecksum(idcard string) bool {
	// 身份证号码必须是18位
	if len(idcard) != 18 {
		return false
	}

	// 前17位必须是数字
	for i := 0; i < 17; i++ {
		if idcard[i] < '0' || idcard[i] > '9' {
			return false
		}
	}

	// 加权因子
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}

	// 校验码对照表
	checkCodes := []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}

	// 计算校验和
	total := 0
	for i := 0; i < 17; i++ {
		digit, _ := strconv.Atoi(string(idcard[i]))
		total += digit * weights[i]
	}

	// 计算模
	mod := total % 11

	// 获取校验码
	expectedCheckCode := checkCodes[mod]

	// 比较校验码（转换为大写比较）
	actualCheckCode := strings.ToUpper(string(idcard[17]))

	return actualCheckCode == expectedCheckCode
}

// VerifyRealName 调用实名认证 API 验证身份证信息
func VerifyRealName(name, idcard string) (*RealNameVerifyResponse, error) {
	// 第一步：使用 ISO 7064:1983, MOD 11-2 算法验证身份证号码
	if !validateIDCardChecksum(idcard) {
		result := &RealNameVerifyResponse{
			Success:         false,
			Message:         "身份证号码格式错误或校验位不正确",
			AliyunAPICalled: false,
		}
		return result, nil
	}

	// 检查是否启用实名认证
	enabled, _ := strconv.ParseBool(os.Getenv("VERIFY_API_ENABLED"))
	if !enabled {
		return nil, fmt.Errorf("real name verification is disabled")
	}

	// 获取阿里云 API 配置
	apiURL := os.Getenv("IDCARD_API_URL")
	if apiURL == "" {
		apiURL = "https://sfzsmyxb.market.alicloudapi.com/get/idcard/checkV3"
	}

	appCode := os.Getenv("IDCARD_APP_CODE")
	if appCode == "" {
		return nil, fmt.Errorf("IDCARD_APP_CODE is not configured")
	}

	// 第二步：调用阿里云 API 进一步验证
	client := &http.Client{Timeout: 30 * time.Second}

	// 准备请求参数
	params := url.Values{}
	params.Add("name", name)
	params.Add("idcard", idcard)

	// 构建请求URL
	requestURL := apiURL + "?" + params.Encode()

	// 创建请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		result := &RealNameVerifyResponse{
			Success:         false,
			Message:         fmt.Sprintf("创建请求失败: %v", err),
			AliyunAPICalled: false,
		}
		return result, err
	}

	// 设置请求头
	req.Header.Set("Authorization", "APPCODE "+appCode)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		result := &RealNameVerifyResponse{
			Success:         false,
			Message:         fmt.Sprintf("请求阿里云 API 失败: %v", err),
			AliyunAPICalled: true,
			APIResponse:     map[string]interface{}{"error": "request_exception", "detail": err.Error()},
		}
		return result, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result := &RealNameVerifyResponse{
			Success:         false,
			Message:         "读取API响应失败",
			AliyunAPICalled: true,
			APIResponse:     map[string]interface{}{"error": "read_response_failed", "detail": err.Error()},
		}
		return result, err
	}

	// 处理HTTP状态码
	if resp.StatusCode != 200 {
		var message string
		switch resp.StatusCode {
		case 400:
			message = "参数错误或 AppCode 错误"
		case 403:
			message = "服务未被授权或套餐包次数用完"
		case 500:
			message = "API 网关错误"
		default:
			message = fmt.Sprintf("请求失败: HTTP %d", resp.StatusCode)
		}

		result := &RealNameVerifyResponse{
			Success:         false,
			Message:         message,
			AliyunAPICalled: true,
			APIResponse:     map[string]interface{}{"http_status": resp.StatusCode, "raw_text": string(body)},
		}
		return result, nil
	}

	// 解析JSON响应
	var apiResp AliyunAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		result := &RealNameVerifyResponse{
			Success:         false,
			Message:         "API 返回数据解析失败",
			AliyunAPICalled: true,
			APIResponse:     map[string]interface{}{"error": "JSON解析失败", "raw_text": string(body)},
		}
		return result, err
	}

	// 将响应转换为map以便保存
	var responseMap map[string]interface{}
	json.Unmarshal(body, &responseMap)

	// 根据阿里云 API 的 code 状态码处理
	var result *RealNameVerifyResponse

	switch apiResp.Code {
	case 200:
		// 成功调用 API，检查验证结果
		data := apiResp.Data
		if data == nil {
			result = &RealNameVerifyResponse{
				Success:         false,
				Message:         "API返回数据为空",
				AliyunAPICalled: true,
				APIResponse:     responseMap,
			}
			break
		}

		resultCode, ok := data["result"].(float64)
		if !ok {
			result = &RealNameVerifyResponse{
				Success:         false,
				Message:         "API返回结果格式错误",
				AliyunAPICalled: true,
				APIResponse:     responseMap,
			}
			break
		}

		orderNo, _ := data["order_no"].(string)
		desc, _ := data["desc"].(string)

		switch int(resultCode) {
		case 0:
			// 身份证信息一致，认证成功
			result = &RealNameVerifyResponse{
				Success:         true,
				Message:         getStringOrDefault(desc, "身份证信息一致"),
				OrderNo:         orderNo,
				AliyunAPICalled: true,
				APIResponse:     responseMap,
			}
		case 1:
			// 身份证信息不一致
			result = &RealNameVerifyResponse{
				Success:         false,
				Message:         getStringOrDefault(desc, "身份证信息不一致"),
				OrderNo:         orderNo,
				AliyunAPICalled: true,
				APIResponse:     responseMap,
			}
		case 2:
			// 无记录
			result = &RealNameVerifyResponse{
				Success:         false,
				Message:         getStringOrDefault(desc, "无记录"),
				OrderNo:         orderNo,
				AliyunAPICalled: true,
				APIResponse:     responseMap,
			}
		default:
			// 未知的 result 值
			result = &RealNameVerifyResponse{
				Success:         false,
				Message:         getStringOrDefault(desc, "验证结果未知"),
				OrderNo:         orderNo,
				AliyunAPICalled: true,
				APIResponse:     responseMap,
			}
		}

	case 400:
		// 参数错误
		result = &RealNameVerifyResponse{
			Success:         false,
			Message:         apiResp.Msg,
			AliyunAPICalled: true,
			APIResponse:     responseMap,
		}

	case 500:
		// 系统异常
		result = &RealNameVerifyResponse{
			Success:         false,
			Message:         getStringOrDefault(apiResp.Msg, "系统异常，请联系服务商"),
			AliyunAPICalled: true,
			APIResponse:     responseMap,
		}

	case 501:
		// 第三方服务异常
		result = &RealNameVerifyResponse{
			Success:         false,
			Message:         getStringOrDefault(apiResp.Msg, "第三方服务异常"),
			AliyunAPICalled: true,
			APIResponse:     responseMap,
		}

	case 604:
		// 接口停用
		result = &RealNameVerifyResponse{
			Success:         false,
			Message:         getStringOrDefault(apiResp.Msg, "接口停用"),
			AliyunAPICalled: true,
			APIResponse:     responseMap,
		}

	default:
		// 其他错误码
		message := getStringOrDefault(apiResp.Msg, fmt.Sprintf("API 返回错误码: %d", apiResp.Code))
		result = &RealNameVerifyResponse{
			Success:         false,
			Message:         message,
			AliyunAPICalled: true,
			APIResponse:     responseMap,
		}
	}

	return result, nil
}

// getStringOrDefault 获取字符串值或默认值
func getStringOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// UpdateUserRealNameStatus 更新用户实名状态并加密存储姓名和身份证号
func UpdateUserRealNameStatus(userId int64, isRealName bool, name, idcard string) error {
	user, err := models.GetUserById(userId)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// 如果实名认证成功，检查该实名信息是否已被其他用户使用
	if isRealName && name != "" && idcard != "" {
		// 检查实名信息唯一性
		isDuplicate, err := checkRealNameDuplicate(userId, name, idcard)
		if err != nil {
			return fmt.Errorf("failed to check real name uniqueness: %v", err)
		}
		if isDuplicate {
			return fmt.Errorf("该实名信息已被其他用户使用")
		}
	}

	// 更新实名状态
	user.IsRealName = isRealName
	user.UpdatedTime = time.Now().Format(time.RFC3339)

	// 如果实名认证成功，加密存储姓名和身份证号
	if isRealName && name != "" && idcard != "" {
		// 加密姓名
		encryptedName, err := EncryptData(name)
		if err != nil {
			return fmt.Errorf("failed to encrypt name: %v", err)
		}
		user.RealName = encryptedName

		// 加密身份证号
		encryptedIDCard, err := EncryptData(idcard)
		if err != nil {
			return fmt.Errorf("failed to encrypt ID card: %v", err)
		}
		user.IDCard = encryptedIDCard
	}

	_, err = models.UpdateUser(userId, user)
	if err != nil {
		return err
	}

	return nil
}

// checkRealNameDuplicate 检查实名信息是否已被其他用户使用
func checkRealNameDuplicate(currentUserId int64, name, idcard string) (bool, error) {
	// 获取所有已实名认证的用户
	var users []models.User
	err := models.GetEngine().Where("is_real_name = ?", true).Find(&users)
	if err != nil {
		return false, err
	}

	// 遍历所有已实名用户，解密并比对
	for _, u := range users {
		// 跳过当前用户
		if u.Id == currentUserId {
			continue
		}

		// 解密姓名
		if u.RealName != "" {
			decryptedName, err := DecryptData(u.RealName)
			if err != nil {
				// 解密失败，记录日志但继续
				fmt.Printf("Warning: Failed to decrypt name for user %d: %v\n", u.Id, err)
				continue
			}

			// 解密身份证号
			if u.IDCard != "" {
				decryptedIDCard, err := DecryptData(u.IDCard)
				if err != nil {
					// 解密失败，记录日志但继续
					fmt.Printf("Warning: Failed to decrypt ID card for user %d: %v\n", u.Id, err)
					continue
				}

				// 比对姓名和身份证号
				if decryptedName == name && decryptedIDCard == idcard {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// GetDecryptedRealName 获取解密后的真实姓名（仅供管理员使用）
func SendVerificationCodeForRealName(target, codeType string) error {
	// 检查是否启用实名认证
	enabled, _ := strconv.ParseBool(os.Getenv("VERIFY_API_ENABLED"))
	if !enabled {
		return fmt.Errorf("real name verification is disabled")
	}

	// 获取 API 地址
	apiUrl := os.Getenv("VERIFY_API_URL")
	if apiUrl == "" {
		apiUrl = "http://localhost:3000"
	}

	// 构建请求
	var url string
	var reqBody map[string]string

	if codeType == "sms" {
		url = fmt.Sprintf("%s/api/sms", apiUrl)
		reqBody = map[string]string{"phone": target}
	} else if codeType == "email" {
		url = fmt.Sprintf("%s/api/email", apiUrl)
		reqBody = map[string]string{"email": target}
	} else {
		return fmt.Errorf("invalid code type: %s", codeType)
	}

	// 序列化请求体
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送 POST 请求
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to call verification API: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// 检查是否成功
	if success, ok := result["success"].(bool); !ok || !success {
		message := "发送失败"
		if msg, ok := result["message"].(string); ok {
			message = msg
		}
		return fmt.Errorf(message)
	}

	return nil
}

// VerifyCodeForRealName 验证实名认证验证码
func VerifyCodeForRealName(target, code, codeType string) (bool, error) {
	// 检查是否启用实名认证
	enabled, _ := strconv.ParseBool(os.Getenv("VERIFY_API_ENABLED"))
	if !enabled {
		return false, fmt.Errorf("real name verification is disabled")
	}

	// 获取 API 地址
	apiUrl := os.Getenv("VERIFY_API_URL")
	if apiUrl == "" {
		apiUrl = "http://localhost:3000"
	}

	// 构建请求
	url := fmt.Sprintf("%s/api/verify", apiUrl)
	reqBody := make(map[string]string)
	reqBody["code"] = code

	if codeType == "sms" {
		reqBody["phone"] = target
	} else if codeType == "email" {
		reqBody["email"] = target
	} else {
		return false, fmt.Errorf("invalid code type: %s", codeType)
	}

	// 序列化请求体
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, err
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送 POST 请求
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to call verification API: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to parse response: %v", err)
	}

	// 检查是否成功
	if success, ok := result["success"].(bool); ok && success {
		return true, nil
	}

	return false, nil
}

// GetDecryptedRealName 获取解密后的真实姓名（仅供管理员使用）
func GetDecryptedRealName(userId int64) (string, error) {
	user, err := models.GetUserById(userId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", fmt.Errorf("user not found")
	}

	if user.RealName == "" {
		return "", nil
	}

	// 解密姓名
	decryptedName, err := DecryptData(user.RealName)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt name: %v", err)
	}

	return decryptedName, nil
}

// GetDecryptedIDCard 获取解密后的身份证号（仅供管理员使用）
func GetDecryptedIDCard(userId int64) (string, error) {
	user, err := models.GetUserById(userId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", fmt.Errorf("user not found")
	}

	if user.IDCard == "" {
		return "", nil
	}

	// 解密身份证号
	decryptedIDCard, err := DecryptData(user.IDCard)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt ID card: %v", err)
	}

	return decryptedIDCard, nil
}

// GetDecryptedRealNameInfo 获取解密后的完整实名信息（仅供管理员使用）
func GetDecryptedRealNameInfo(userId int64) (name, idcard string, err error) {
	user, err := models.GetUserById(userId)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", fmt.Errorf("user not found")
	}

	// 解密姓名
	if user.RealName != "" {
		name, err = DecryptData(user.RealName)
		if err != nil {
			return "", "", fmt.Errorf("failed to decrypt name: %v", err)
		}
	}

	// 解密身份证号
	if user.IDCard != "" {
		idcard, err = DecryptData(user.IDCard)
		if err != nil {
			return "", "", fmt.Errorf("failed to decrypt ID card: %v", err)
		}
	}

	return name, idcard, nil
}
