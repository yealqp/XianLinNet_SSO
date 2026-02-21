package types

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
	Email    string `json:"email" validate:"required,email"`
	QQ       string `json:"qq,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	QQ       string `json:"qq,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// CreateApplicationRequest 创建应用请求
type CreateApplicationRequest struct {
	Name         string   `json:"name" validate:"required"`
	DisplayName  string   `json:"displayName,omitempty"`
	Logo         string   `json:"logo,omitempty"`
	Organization string   `json:"organization,omitempty"`
	RedirectUris []string `json:"redirectUris,omitempty"`
	GrantTypes   []string `json:"grantTypes,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}
