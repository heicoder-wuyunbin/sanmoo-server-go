package handler

import (
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"sanmoo-server-go/internal/shared/netutil"
	"sanmoo-server-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ======================== Auth Handlers ========================

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.Auth.Login(c.Request.Context(), req.Username, req.Password, req.Code, netutil.RealIP(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

// CheckMFA 检查指定用户是否需要邮箱验证码（无需密码）
func (h *Handler) CheckMFA(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	needMFA, err := h.svc.Auth.CheckMFA(c.Request.Context(), req.Username)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, gin.H{"needMfa": needMFA})
}

func (h *Handler) SendLoginVerificationCode(c *gin.Context) {
	var req dto.SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.Auth.SendLoginVerificationCode(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) VerifyLoginVerificationCode(c *gin.Context) {
	var req dto.VerifyVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.Auth.VerifyLoginVerificationCode(c.Request.Context(), req.UserID, req.Code)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperr.ErrInvalidParam)
		return
	}
	out, err := h.svc.Auth.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Ok(c, out)
}
