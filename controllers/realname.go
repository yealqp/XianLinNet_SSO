package controllers

import (
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
)

type RealNameController struct {
	BaseController
}

// VerifyRealName 验证实名信息
func (c *RealNameController) VerifyRealName() {
	var req struct {
		Name   string `json:"name"`
		IDCard string `json:"idcard"`
	}

	err := c.GetRequestBody(&req)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	if req.Name == "" || req.IDCard == "" {
		c.ResponseError("姓名和身份证号不能为空")
		return
	}

	result, err := services.VerifyRealName(req.Name, req.IDCard)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !result.Success {
		c.ResponseError(result.Message)
		return
	}

	c.ResponseOk(map[string]interface{}{
		"message":  result.Message,
		"order_no": result.OrderNo,
	})
}

// SubmitRealName 提交实名认证
func (c *RealNameController) SubmitRealName() {
	var req struct {
		UserId int64  `json:"userId"`
		Name   string `json:"name"`
		IDCard string `json:"idcard"`
	}

	err := c.GetRequestBody(&req)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	if req.Name == "" || req.IDCard == "" {
		c.ResponseError("姓名和身份证号不能为空")
		return
	}

	if req.UserId == 0 {
		c.ResponseError("用户ID不能为空")
		return
	}

	user, err := models.GetUserById(req.UserId)
	if err != nil || user == nil {
		c.ResponseError("用户不存在")
		return
	}

	if user.IsRealName {
		c.ResponseError("您已完成实名认证")
		return
	}

	result, err := services.VerifyRealName(req.Name, req.IDCard)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !result.Success {
		c.ResponseError(result.Message)
		return
	}

	err = services.UpdateUserRealNameStatus(req.UserId, true, req.Name, req.IDCard)
	if err != nil {
		c.ResponseError("实名认证成功，但更新状态失败: " + err.Error())
		return
	}

	c.ResponseOk(map[string]interface{}{
		"message":  "实名认证成功",
		"order_no": result.OrderNo,
	})
}

// GetRealNameInfo 获取用户实名信息（仅管理员）
// @router /api/admin/realname/:userId [get]
func (c *RealNameController) GetRealNameInfo() {
	// 检查管理员权限
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Missing authorization header")
		return
	}

	// Extract token
	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Invalid authorization header")
		return
	}

	// Validate token and check admin
	user, err := services.ValidateToken(token)
	if err != nil {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Invalid token")
		return
	}

	if !user.IsAdmin {
		c.Ctx.Output.SetStatus(403)
		c.ResponseError("Admin access required")
		return
	}

	// 获取用户ID
	userIdStr := c.GetString(":userId")
	userId, err := c.GetInt64(":userId")
	if err != nil || userId == 0 {
		c.ResponseError("Invalid user ID: " + userIdStr)
		return
	}

	// 获取解密后的实名信息
	name, idcard, err := services.GetDecryptedRealNameInfo(userId)
	if err != nil {
		c.ResponseError("获取实名信息失败: " + err.Error())
		return
	}

	// 脱敏处理身份证号（只显示前6位和后4位）
	maskedIDCard := ""
	if len(idcard) >= 10 {
		maskedIDCard = idcard[:6] + "********" + idcard[len(idcard)-4:]
	} else {
		maskedIDCard = idcard
	}

	c.ResponseOk(map[string]interface{}{
		"name":   name,
		"idcard": maskedIDCard,
	})
}
