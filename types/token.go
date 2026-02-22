package types

// TokenRequest Token 请求
type TokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type" validate:"required"`
	Code         string `json:"code,omitempty" form:"code"`
	RedirectUri  string `json:"redirect_uri,omitempty" form:"redirect_uri"`
	ClientId     string `json:"client_id,omitempty" form:"client_id"`
	ClientSecret string `json:"client_secret,omitempty" form:"client_secret"`
	RefreshToken string `json:"refresh_token,omitempty" form:"refresh_token"`
	Username     string `json:"username,omitempty" form:"username"`
	Password     string `json:"password,omitempty" form:"password"`
	Scope        string `json:"scope,omitempty" form:"scope"`
}

// TokenResponse Token 响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}
