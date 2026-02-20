// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/oauth-server/oauth-server/models"
)

// RealNameVerifyRequest 实名认证请求
type RealNameVerifyRequest struct {
	Name   string `json:"name"`
	IDCard string `json:"idcard"`
}

// RealNameVerifyResponse 实名认证响应
type RealNameVerifyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	OrderNo string `json:"order_no"`
}

// VerifyRealName 调用实名认证 API 验证身份证信息
func VerifyRealName(name, idcard string) (*RealNameVerifyResponse, error) {
	// 检查是否启用实名认证
	enabled, _ := web.AppConfig.Bool("verifyApiEnabled")
	if !enabled {
		return nil, fmt.Errorf("real name verification is disabled")
	}

	// 获取 API 地址
	apiUrl, _ := web.AppConfig.String("verifyApiUrl")
	if apiUrl == "" {
		apiUrl = "http://localhost:3000"
	}

	// 构建请求 URL
	url := fmt.Sprintf("%s/api/idcard?name=%s&idcard=%s", apiUrl, name, idcard)

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送 GET 请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call verification API: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var result RealNameVerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
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

	// 如果实名认证成功，升级用户角色
	if isRealName {
		err = upgradeUserRole(user)
		if err != nil {
			// 记录错误但不影响实名状态更新
			fmt.Printf("Warning: Failed to upgrade user role: %v\n", err)
		}
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

// upgradeUserRole 升级用户角色（从未实名用户到普通用户）
func upgradeUserRole(user *models.User) error {
	// 移除未实名用户角色
	_, err := models.RemoveUserRole(user.Id, "admin", "unverified-user")
	if err != nil {
		// Log error but continue
		fmt.Printf("Warning: Failed to remove unverified-user role: %v\n", err)
	}

	// 添加普通用户角色
	_, err = models.AddUserRole(user.Id, "admin", "normal-user")
	if err != nil {
		return err
	}

	return nil
}

// SendVerificationCodeForRealName 发送实名认证验证码
func SendVerificationCodeForRealName(target, codeType string) error {
	// 检查是否启用实名认证
	enabled, _ := web.AppConfig.Bool("verifyApiEnabled")
	if !enabled {
		return fmt.Errorf("real name verification is disabled")
	}

	// 获取 API 地址
	apiUrl, _ := web.AppConfig.String("verifyApiUrl")
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
	enabled, _ := web.AppConfig.Bool("verifyApiEnabled")
	if !enabled {
		return false, fmt.Errorf("real name verification is disabled")
	}

	// 获取 API 地址
	apiUrl, _ := web.AppConfig.String("verifyApiUrl")
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
