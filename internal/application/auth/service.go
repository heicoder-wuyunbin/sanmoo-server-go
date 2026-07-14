package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	domuser "sanmoo-server-go/internal/domain/user"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/infrastructure/email"
	"sanmoo-server-go/internal/infrastructure/security"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	userRepo            domuser.Repository
	jwt                 *security.JWTManager
	verificationService *cache.VerificationService
	emailService        *email.EmailService
	wxAppID             string
	wxSecret            string
	httpClient          *http.Client
}

func NewService(
	userRepo domuser.Repository,
	jwt *security.JWTManager,
	vs *cache.VerificationService,
	emailService *email.EmailService,
	wxAppID string,
	wxSecret string,
) *Service {
	return &Service{
		userRepo:            userRepo,
		jwt:                 jwt,
		verificationService: vs,
		emailService:        emailService,
		wxAppID:             wxAppID,
		wxSecret:            wxSecret,
		httpClient:          &http.Client{Timeout: 8 * time.Second},
	}
}

// Login 后台登录：
// - 如果开启邮箱验证码（emailConfig.loginMfaEnabled=true），则必须携带 code；
// - code 通过 /auth/send-verification-code 下发到用户绑定邮箱。
func (s *Service) Login(ctx context.Context, username, password, code, clientIP string) (any, error) {
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, apperr.ErrBadCredential
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, apperr.ErrBadCredential
	}

	// 开启邮箱验证码后，后台管理员登录必须校验验证码
	if s.emailService != nil && s.emailService.IsMFAEnabled() && strings.Contains(strings.ToUpper(u.RoleName), "ADMIN") {
		if code == "" {
			return nil, apperr.ErrMFARequired
		}
		key := cache.GenerateLoginVerificationKey(u.ID)
		stored, err := s.verificationService.GetCode(ctx, key)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil, apperr.ErrBadVerifyCode
			}
			return nil, err
		}
		if stored != code {
			return nil, apperr.ErrBadVerifyCode
		}
		_ = s.verificationService.DeleteCode(ctx, key)
	}

	// 记录真实 IP（即使记录失败也不影响登录）
	now := time.Now()
	u.LastLoginTime = &now
	u.LastLoginIp = clientIP
	_ = s.userRepo.UpdateUser(ctx, u)

	accessToken, _ := s.jwt.GenerateAccessToken(u.ID, u.Username, u.RoleID, "admin")
	refreshToken, _ := s.jwt.GenerateRefreshToken(u.ID, u.Username, u.RoleID, "admin")
	return &dto.AuthLoginResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: u}, nil
}

func (s *Service) SendLoginVerificationCode(ctx context.Context, username, password string) (*dto.SendVerificationCodeResponse, error) {
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, apperr.ErrBadCredential
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, apperr.ErrBadCredential
	}

	if s.emailService == nil || !s.emailService.IsConfigured() {
		return nil, apperr.New(apperr.ErrInvalidParam.Code, "邮件服务未配置")
	}
	if !s.emailService.IsMFAEnabled() || !strings.Contains(strings.ToUpper(u.RoleName), "ADMIN") {
		return nil, apperr.New(apperr.ErrInvalidParam.Code, "当前未开启后台登录邮箱验证码")
	}
	if u.Email == "" {
		return nil, apperr.New(apperr.ErrInvalidParam.Code, "账号未绑定邮箱")
	}

	code, err := s.verificationService.GenerateCode()
	if err != nil {
		return nil, err
	}
	identifier, err := s.verificationService.GenerateIdentifier()
	if err != nil {
		return nil, err
	}
	key := cache.GenerateLoginVerificationKey(u.ID)
	if err := s.verificationService.StoreCode(ctx, key, code); err != nil {
		return nil, err
	}
	if err := s.emailService.SendVerificationCodeWithIdentifier(u.Email, code, identifier); err != nil {
		return nil, err
	}
	// 前端依赖 userId 进行二次校验
	return &dto.SendVerificationCodeResponse{UserID: u.ID, Identifier: identifier}, nil
}

// CheckMFA 检查指定用户是否需要邮箱验证码（无需密码验证）
func (s *Service) CheckMFA(ctx context.Context, username string) (bool, error) {
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return false, apperr.ErrBadCredential
	}
	if s.emailService == nil || !s.emailService.IsConfigured() {
		return false, nil
	}
	if !s.emailService.IsMFAEnabled() {
		return false, nil
	}
	if !strings.Contains(strings.ToUpper(u.RoleName), "ADMIN") {
		return false, nil
	}
	return u.Email != "", nil
}

func (s *Service) VerifyLoginVerificationCode(ctx context.Context, userID uint64, code string) (*dto.AuthLoginResponse, error) {
	if userID == 0 || code == "" {
		return nil, apperr.ErrInvalidParam
	}
	key := cache.GenerateLoginVerificationKey(userID)
	stored, err := s.verificationService.GetCode(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperr.ErrBadVerifyCode
		}
		return nil, err
	}
	if stored != code {
		return nil, apperr.ErrBadVerifyCode
	}
	_ = s.verificationService.DeleteCode(ctx, key)

	u, err := s.userRepo.FindByIDUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	accessToken, _ := s.jwt.GenerateAccessToken(u.ID, u.Username, u.RoleID, "admin")
	refreshToken, _ := s.jwt.GenerateRefreshToken(u.ID, u.Username, u.RoleID, "admin")
	return &dto.AuthLoginResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: u}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthRefreshResponse, error) {
	claims, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	accessToken, _ := s.jwt.GenerateAccessToken(claims.UserID, claims.Username, claims.RoleID, claims.RoleName)
	newRefreshToken, _ := s.jwt.GenerateRefreshToken(claims.UserID, claims.Username, claims.RoleID, claims.RoleName)
	return &dto.AuthRefreshResponse{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}

func (s *Service) MPAuthSession(ctx context.Context, code string) (*dto.MPAuthSessionResponse, error) {
	if code == "" {
		return nil, apperr.ErrInvalidParam
	}

	// 调用微信 jscode2session 获取 openid
	type resp struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if s.wxAppID == "" || s.wxSecret == "" {
		return nil, apperr.New(apperr.ErrInvalidParam.Code, "微信小程序 AppID/Secret 未配置")
	}
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.wxAppID, s.wxSecret, code,
	)
	r, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	var out resp
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.ErrCode != 0 || out.OpenID == "" {
		return nil, apperr.New(apperr.ErrInvalidParam.Code, fmt.Sprintf("微信登录失败: %s", out.ErrMsg))
	}
	openID := out.OpenID
	if err := s.userRepo.UpsertMPUser(ctx, openID); err != nil {
		return nil, err
	}
	return &dto.MPAuthSessionResponse{OpenID: openID}, nil
}
