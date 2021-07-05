package auth

type LoginInfo struct {
	Account  string `form:"account" json:"account" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type TokenInfo struct {
	AccessToken       string `json:"access_token,omitempty"`
	AuthorizationType string `json:"authorization_type,omitempty"`
	ExpiresIn         int64  `json:"expires_in,omitempty"`
}
