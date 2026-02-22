// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
	"github.com/oauth-server/oauth-server/types"
)

// SubmitRealNameRequest 提交实名认证请求
type SubmitRealNameRequest struct {
	Name   string `json:"name"`
	IDCard string `json:"idcard"`
}

// HandleSubmitRealName 处理提交实名认证信息请求
// Requirements: 9.1, 9.4
func HandleSubmitRealName() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 需要 JWT 认证，从 context 获取用户 ID
		userID, ok := ctx.Locals("userID").(string)
		if !ok || userID == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("未授权"))
		}

		// 解析请求
		var req SubmitRealNameRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证参数
		if req.Name == "" || req.IDCard == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("姓名和身份证号不能为空"))
		}

		// 转换 userID 为 int64
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		// 获取用户
		user, err := models.GetUserById(userIDInt)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 检查用户是否已实名认证
		if user.IsRealName {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户已完成实名认证"))
		}

		// 调用实名认证 API 验证身份证信息
		verifyResult, err := services.VerifyRealName(req.Name, req.IDCard)
		if err != nil {
			// API 调用失败，返回错误信息
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("实名认证服务异常: " + err.Error()))
		}

		// 检查验证结果
		if !verifyResult.Success {
			// 验证失败，返回失败信息
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse(verifyResult.Message))
		}

		// 验证成功，更新用户实名状态并加密存储姓名和身份证号
		err = services.UpdateUserRealNameStatus(userIDInt, true, req.Name, req.IDCard)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse(err.Error()))
		}

		// 返回成功响应
		return ctx.JSON(types.SuccessResponse(map[string]interface{}{
			"message":    "实名认证成功",
			"isRealName": true,
			"order_no":   verifyResult.OrderNo,
		}))
	}
}

// HandleGetRealNameInfo 处理获取实名认证状态请求
// Requirements: 9.2
func HandleGetRealNameInfo() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 需要 JWT 认证，从 context 获取用户 ID
		userID, ok := ctx.Locals("userID").(string)
		if !ok || userID == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("未授权"))
		}

		// 转换 userID 为 int64
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		// 获取用户
		user, err := models.GetUserById(userIDInt)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 构造响应数据（不返回敏感信息）
		realNameInfo := map[string]interface{}{
			"isRealName": user.IsRealName,
		}

		// 如果用户已实名认证，返回脱敏的信息
		if user.IsRealName {
			// 解密姓名和身份证号（仅用于脱敏显示）
			name, idcard, err := services.GetDecryptedRealNameInfo(userIDInt)
			if err != nil {
				// 解密失败，仅返回实名状态
				return ctx.JSON(types.SuccessResponse(realNameInfo))
			}

			// 脱敏处理：姓名显示第一个字，其余用*代替
			if len(name) > 0 {
				maskedName := string([]rune(name)[0]) + "**"
				realNameInfo["name"] = maskedName
			}

			// 脱敏处理：身份证号显示前4位和后2位，中间用*代替
			if len(idcard) >= 10 {
				maskedIDCard := idcard[:4] + "************" + idcard[len(idcard)-2:]
				realNameInfo["idcard"] = maskedIDCard
			}
		}

		return ctx.JSON(types.SuccessResponse(realNameInfo))
	}
}

// HandleAdminGetRealNameInfo 处理管理员获取用户实名信息请求
// Requirements: 9.2
func HandleAdminGetRealNameInfo() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 需要管理员权限（已通过 AdminAuthMiddleware 验证）

		// 从路径参数获取用户 ID
		userIDStr := ctx.Params("userId")
		if userIDStr == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户ID不能为空"))
		}

		// 转换 userID 为 int64
		userIDInt, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		// 获取用户
		user, err := models.GetUserById(userIDInt)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 如果用户未实名认证，返回空信息
		if !user.IsRealName {
			return ctx.JSON(types.SuccessResponse(map[string]interface{}{
				"isRealName": false,
				"name":       "",
				"idcard":     "",
			}))
		}

		// 获取解密后的实名信息（管理员可查看完整信息）
		name, idcard, err := services.GetDecryptedRealNameInfo(userIDInt)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取实名信息失败"))
		}

		// 返回完整的实名信息供管理员查看
		return ctx.JSON(types.SuccessResponse(map[string]interface{}{
			"isRealName": true,
			"name":       name,
			"idcard":     idcard,
		}))
	}
}

// VerifyRealNameRequest 管理员验证实名信息请求
type VerifyRealNameRequest struct {
	UserID     int64 `json:"userId"`
	IsApproved bool  `json:"isApproved"`
}

// HandleVerifyRealName 处理管理员验证实名信息请求
// Requirements: 9.3
func HandleVerifyRealName() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 需要管理员权限（已通过 AdminAuthMiddleware 验证）

		// 解析请求
		var req VerifyRealNameRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证参数
		if req.UserID == 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户ID不能为空"))
		}

		// 获取用户
		user, err := models.GetUserById(req.UserID)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 获取解密后的实名信息（仅供管理员查看）
		name, idcard, err := services.GetDecryptedRealNameInfo(req.UserID)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取实名信息失败"))
		}

		// 如果管理员拒绝实名认证，清除实名信息
		if !req.IsApproved {
			err = services.UpdateUserRealNameStatus(req.UserID, false, "", "")
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("更新实名状态失败"))
			}

			return ctx.JSON(types.SuccessResponse(map[string]interface{}{
				"message":    "实名认证已拒绝",
				"isRealName": false,
			}))
		}

		// 管理员批准实名认证（实名状态已经是 true，无需更新）
		// 返回完整的实名信息供管理员查看
		return ctx.JSON(types.SuccessResponse(map[string]interface{}{
			"message":    "实名认证已批准",
			"isRealName": true,
			"userId":     user.Id,
			"username":   user.Username,
			"email":      user.Email,
			"name":       name,
			"idcard":     idcard,
		}))
	}
}
