package types

// LoginRequest 登录请求
type LoginRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6"`
	CaptchaToken string `json:"captchaToken,omitempty"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	ExpiresIn    int      `json:"expires_in"`
	TokenType    string   `json:"token_type"`
	User         UserInfo `json:"user"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username         string `json:"username" validate:"required,min=3,max=50"`
	Password         string `json:"password" validate:"required,min=6"`
	Email            string `json:"email" validate:"required,email"`
	VerificationCode string `json:"verificationCode" validate:"required"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	IsAdmin     bool   `json:"isAdmin"`
	IsRealName  bool   `json:"isRealName,omitempty"`
	IsForbidden bool   `json:"isForbidden"`
	QQ          string `json:"qq,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
}
