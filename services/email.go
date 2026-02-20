// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math/big"
	"net/smtp"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// EmailVerificationCode stores verification code information
type EmailVerificationCode struct {
	Email       string
	Code        string
	Purpose     string // "register", "reset_password"
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Attempts    int
	MaxAttempts int
}

// In-memory storage for verification codes (in production, use Redis)
var verificationCodes = make(map[string]*EmailVerificationCode)

// GenerateVerificationCode generates a 6-digit verification code using crypto/rand
func GenerateVerificationCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		// Generate cryptographically secure random number between 0-9
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			// Fallback to time-based if crypto/rand fails
			n = big.NewInt(time.Now().UnixNano() % 10)
		}
		code += n.String()
	}
	return code
}

// SendEmailViaSMTP sends email using SMTP with SSL/TLS support
func SendEmailViaSMTP(to, subject, body string) error {
	smtpEnabled, _ := web.AppConfig.Bool("smtpEnabled")
	if !smtpEnabled {
		logs.Info("SMTP disabled. Email to %s", to)
		return nil
	}

	smtpHost, _ := web.AppConfig.String("smtpHost")
	smtpPort, _ := web.AppConfig.String("smtpPort")
	smtpUseSSL, _ := web.AppConfig.Bool("smtpUseSSL")
	smtpUser, _ := web.AppConfig.String("smtpUser")
	smtpPassword, _ := web.AppConfig.String("smtpPassword")
	smtpFrom, _ := web.AppConfig.String("smtpFrom")
	smtpFromName, _ := web.AppConfig.String("smtpFromName")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPassword == "" {
		return fmt.Errorf("SMTP configuration incomplete")
	}

	// Setup authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)

	// Compose message
	from := fmt.Sprintf("%s <%s>", smtpFromName, smtpFrom)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", from, to, subject, body))

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	// Send email with SSL/TLS support
	if smtpUseSSL {
		// Use SSL (port 465)
		tlsConfig := &tls.Config{
			ServerName: smtpHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			logs.Error("Failed to connect with SSL: %v", err)
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, smtpHost)
		if err != nil {
			logs.Error("Failed to create SMTP client: %v", err)
			return err
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			logs.Error("Failed to authenticate: %v", err)
			return err
		}

		if err = client.Mail(smtpFrom); err != nil {
			return err
		}

		if err = client.Rcpt(to); err != nil {
			return err
		}

		w, err := client.Data()
		if err != nil {
			return err
		}

		_, err = w.Write(msg)
		if err != nil {
			return err
		}

		err = w.Close()
		if err != nil {
			return err
		}

		client.Quit()
	} else {
		// Use STARTTLS (port 587)
		err := smtp.SendMail(addr, auth, smtpFrom, []string{to}, msg)
		if err != nil {
			logs.Error("Failed to send email: %v", err)
			return err
		}
	}

	logs.Info("Email sent successfully to %s", to)
	return nil
}

// SendVerificationEmail sends verification code via email
func SendVerificationEmail(email, purpose string) (string, error) {
	// Generate verification code
	code := GenerateVerificationCode()

	// Store verification code
	key := fmt.Sprintf("%s:%s", email, purpose)
	verificationCodes[key] = &EmailVerificationCode{
		Email:       email,
		Code:        code,
		Purpose:     purpose,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute), // 10 minutes expiration
		Attempts:    0,
		MaxAttempts: 5,
	}

	// Prepare email content
	var subject, body string
	if purpose == "register" {
		subject = "XianlinNet ID - 注册验证码"
		body = getRegisterEmailTemplate(code)
	} else if purpose == "reset_password" {
		subject = "XianlinNet ID - 密码重置验证码"
		body = getResetPasswordEmailTemplate(code)
	}

	// Send email
	err := SendEmailViaSMTP(email, subject, body)
	if err != nil {
		logs.Error("Failed to send verification email: %v", err)
		// In development, still log the code
		logs.Info("Verification code for %s (%s): %s", email, purpose, code)
	}

	// In development, return the code for testing
	smtpEnabled, _ := web.AppConfig.Bool("smtpEnabled")
	if !smtpEnabled {
		logs.Info("Verification code for %s (%s): %s", email, purpose, code)
		return code, nil
	}

	return "", nil
}

// getRegisterEmailTemplate returns the HTML template for registration email
func getRegisterEmailTemplate(code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>注册验证码</title>
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f7fa;">
    <table width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color: #f5f7fa; padding: 40px 0;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" border="0" style="background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); overflow: hidden;">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #f6339a 0%%, #ff4db3 100%%); padding: 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: 600;">XianlinNet ID</h1>
                            <p style="margin: 10px 0 0 0; color: rgba(255,255,255,0.9); font-size: 14px;">安全、可靠的身份认证服务</p>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="margin: 0 0 20px 0; color: #333333; font-size: 22px; font-weight: 600;">欢迎注册 XianlinNet ID</h2>
                            <p style="margin: 0 0 30px 0; color: #666666; font-size: 15px; line-height: 1.6;">
                                感谢您选择 XianlinNet ID。为了确保账号安全，请使用以下验证码完成注册：
                            </p>
                            
                            <!-- Verification Code Box -->
                            <table width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin: 30px 0;">
                                <tr>
                                    <td align="center" style="background: linear-gradient(135deg, #f6339a 0%%, #ff4db3 100%%); border-radius: 8px; padding: 30px;">
                                        <div style="font-size: 36px; font-weight: bold; color: #ffffff; letter-spacing: 8px; font-family: 'Courier New', monospace;">
                                            %s
                                        </div>
                                    </td>
                                </tr>
                            </table>
                            
                            <div style="background-color: #f8f9fa; border-left: 4px solid #f6339a; padding: 15px 20px; margin: 30px 0; border-radius: 4px;">
                                <p style="margin: 0; color: #666666; font-size: 14px; line-height: 1.6;">
                                    <strong style="color: #333333;">重要提示：</strong><br>
                                    • 验证码有效期为 <strong>10 分钟</strong><br>
                                    • 请勿将验证码告知他人<br>
                                    • 如非本人操作，请忽略此邮件
                                </p>
                            </div>
                            
                            <p style="margin: 30px 0 0 0; color: #999999; font-size: 13px; line-height: 1.6;">
                                如果您没有请求此验证码，可能是他人误输入了您的邮箱地址。您可以放心地忽略此邮件。
                            </p>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8f9fa; padding: 30px; text-align: center; border-top: 1px solid #e9ecef;">
                            <p style="margin: 0 0 10px 0; color: #999999; font-size: 12px;">
                                此邮件由系统自动发送，请勿直接回复
                            </p>
                            <p style="margin: 0; color: #999999; font-size: 12px;">
                                &copy; 2024 XianlinNet. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, code)
}

// getResetPasswordEmailTemplate returns the HTML template for password reset email
func getResetPasswordEmailTemplate(code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>密码重置验证码</title>
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f7fa;">
    <table width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color: #f5f7fa; padding: 40px 0;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" border="0" style="background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); overflow: hidden;">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #f6339a 0%%, #ff4db3 100%%); padding: 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: 600;">XianlinNet ID</h1>
                            <p style="margin: 10px 0 0 0; color: rgba(255,255,255,0.9); font-size: 14px;">安全、可靠的身份认证服务</p>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="margin: 0 0 20px 0; color: #333333; font-size: 22px; font-weight: 600;">密码重置请求</h2>
                            <p style="margin: 0 0 30px 0; color: #666666; font-size: 15px; line-height: 1.6;">
                                我们收到了您的密码重置请求。请使用以下验证码完成密码重置：
                            </p>
                            
                            <!-- Verification Code Box -->
                            <table width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin: 30px 0;">
                                <tr>
                                    <td align="center" style="background: linear-gradient(135deg, #f6339a 0%%, #ff4db3 100%%); border-radius: 8px; padding: 30px;">
                                        <div style="font-size: 36px; font-weight: bold; color: #ffffff; letter-spacing: 8px; font-family: 'Courier New', monospace;">
                                            %s
                                        </div>
                                    </td>
                                </tr>
                            </table>
                            
                            <div style="background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px 20px; margin: 30px 0; border-radius: 4px;">
                                <p style="margin: 0; color: #856404; font-size: 14px; line-height: 1.6;">
                                    <strong style="color: #856404;">⚠️ 安全警告</strong><br>
                                    • 验证码有效期为 <strong>10 分钟</strong><br>
                                    • 如果这不是您的操作，请立即更改密码<br>
                                    • 建议启用两步验证以提高账号安全性
                                </p>
                            </div>
                            
                            <p style="margin: 30px 0 0 0; color: #999999; font-size: 13px; line-height: 1.6;">
                                如果您没有请求重置密码，请忽略此邮件。您的密码不会被更改。
                            </p>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8f9fa; padding: 30px; text-align: center; border-top: 1px solid #e9ecef;">
                            <p style="margin: 0 0 10px 0; color: #999999; font-size: 12px;">
                                此邮件由系统自动发送，请勿直接回复
                            </p>
                            <p style="margin: 0; color: #999999; font-size: 12px;">
                                &copy; 2024 XianlinNet. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, code)
}

// VerifyCode verifies the email verification code
func VerifyCode(email, code, purpose string) (bool, error) {
	key := fmt.Sprintf("%s:%s", email, purpose)

	vc, exists := verificationCodes[key]
	if !exists {
		return false, fmt.Errorf("verification code not found")
	}

	// Check if expired
	if time.Now().After(vc.ExpiresAt) {
		delete(verificationCodes, key)
		return false, fmt.Errorf("verification code expired")
	}

	// Check attempts
	if vc.Attempts >= vc.MaxAttempts {
		delete(verificationCodes, key)
		return false, fmt.Errorf("too many attempts")
	}

	// Increment attempts
	vc.Attempts++

	// Verify code
	if vc.Code != code {
		return false, fmt.Errorf("invalid verification code")
	}

	// Code is valid, delete it
	delete(verificationCodes, key)
	return true, nil
}

// CleanupExpiredCodes removes expired verification codes
func CleanupExpiredCodes() {
	for key, vc := range verificationCodes {
		if time.Now().After(vc.ExpiresAt) {
			delete(verificationCodes, key)
		}
	}
}

// Initialize cleanup routine
func init() {
	// Run cleanup every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			CleanupExpiredCodes()
		}
	}()
}
