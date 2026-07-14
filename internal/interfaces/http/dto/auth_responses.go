package dto

import domuser "sanmoo-server-go/internal/domain/user"

type AuthLoginResponse struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	User         *domuser.User `json:"user"`
}

type AuthRefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// AuthMFARequiredResponse 用于后台开启邮箱验证码登录时的提示响应。
type AuthMFARequiredResponse struct {
	RequiresVerification bool   `json:"requiresVerification"`
	UserID               uint64 `json:"userId"`
}

type MPAuthSessionResponse struct {
	OpenID string `json:"openid"`
}

type SendVerificationCodeResponse struct {
	UserID     uint64 `json:"userId"`
	Identifier string `json:"identifier"`
}
