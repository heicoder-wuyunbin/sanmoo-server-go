package setting

import (
	"context"
	"encoding/json"

	domsetting "sanmoo-server-go/internal/domain/setting"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/infrastructure/email"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
	"strings"
	"time"
)

type Service struct {
	repo      domsetting.Repository
	emailServ *email.EmailService
	verifySvc *cache.VerificationService
	bizCache  *cache.BusinessCache
}

func NewService(repo domsetting.Repository, emailServ *email.EmailService, verifySvc *cache.VerificationService, bizCache *cache.BusinessCache) *Service {
	return &Service{repo: repo, emailServ: emailServ, verifySvc: verifySvc, bizCache: bizCache}
}

func (s *Service) GetSettings(ctx context.Context) (*dto.SettingsResponse, error) {
	st, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.SettingsResponse{
		CoreConfig:    st.CoreConfig,
		UIConfig:      st.UIConfig,
		StorageConfig: st.StorageConfig,
		EmailConfig:   st.EmailConfig,
	}, nil
}

func (s *Service) UpdateSettings(ctx context.Context, body map[string]any, operator string) error {
	st, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}

	// --- 邮箱配置：要求“先验证后保存” ---
	// 若请求中携带 emailConfig，且与当前配置不一致，则必须先通过验证码验证。
	if newEmailCfg, ok := body["emailConfig"].(map[string]any); ok {
		newUsername, _ := newEmailCfg["username"].(string)
		newUsername = strings.TrimSpace(newUsername)
		if newUsername != "" {
			needVerify := true
			if st.EmailConfig != nil && len(st.EmailConfig) > 0 {
				// 只比较关键字段即可
				old := st.EmailConfig
				get := func(m map[string]any, k string) string {
					if v, ok := m[k].(string); ok {
						return strings.TrimSpace(v)
					}
					return ""
				}
				needVerify =
					get(old, "host") != get(newEmailCfg, "host") ||
						get(old, "port") != get(newEmailCfg, "port") ||
						get(old, "username") != get(newEmailCfg, "username") ||
						get(old, "password") != get(newEmailCfg, "password") ||
						get(old, "from") != get(newEmailCfg, "from")
			}
			if needVerify {
				if s.verifySvc == nil {
					return apperr.New(apperr.ErrInternal.Code, "验证码服务未初始化")
				}
				markKey := cache.GenerateEmailVerifiedKey(newUsername)
				mark, err := s.verifySvc.GetCode(ctx, markKey)
				if err != nil || mark != "1" {
					return apperr.ErrEmailNotVerified
				}
				// 允许一次性使用，保存成功后会清理该标记
			}
		}
	}

	// 只更新请求中出现的配置块，避免误覆盖其它配置
	if v, ok := body["coreConfig"].(map[string]any); ok {
		st.CoreConfig = v
	}
	if v, ok := body["uiConfig"].(map[string]any); ok {
		st.UIConfig = v
	}
	if v, ok := body["storageConfig"].(map[string]any); ok {
		st.StorageConfig = v
	}
	if v, ok := body["emailConfig"].(map[string]any); ok {
		st.EmailConfig = v
	}

	if err := s.repo.Update(ctx, st, operator); err != nil {
		return err
	}

	// 保存成功后，清理邮箱“已验证”标记（一次性使用）
	if newEmailCfg, ok := body["emailConfig"].(map[string]any); ok {
		if u, ok := newEmailCfg["username"].(string); ok && strings.TrimSpace(u) != "" && s.verifySvc != nil {
			_ = s.verifySvc.DeleteCode(ctx, cache.GenerateEmailVerifiedKey(strings.TrimSpace(u)))
		}
	}

	// 同步更新内存中的邮件配置（用于验证码等发送能力）
	if s.emailServ != nil && st.EmailConfig != nil {
		s.emailServ.UpdateConfig(st.EmailConfig)
	}
	// 清除设置相关的缓存
	if s.bizCache != nil {
		_ = s.bizCache.DeletePattern(ctx, "blog:setting:*")
	}
	return nil
}

// SendEmailVerificationCode 给当前 emailConfig.username 发送验证码（不落库），并写入 Redis。
// 用于“绑定邮箱前先验证”的流程。
func (s *Service) SendEmailVerificationCode(ctx context.Context, cfg dto.EmailConfigRequest) (string, error) {
	if s.emailServ == nil {
		return "", apperr.New(apperr.ErrInvalidParam.Code, "邮件服务未配置")
	}
	if s.verifySvc == nil {
		return "", apperr.New(apperr.ErrInternal.Code, "验证服务未初始化")
	}
	to := strings.TrimSpace(cfg.Username)
	if to == "" || strings.TrimSpace(cfg.Host) == "" || strings.TrimSpace(cfg.Port) == "" ||
		strings.TrimSpace(cfg.Password) == "" || strings.TrimSpace(cfg.From) == "" {
		return "", apperr.ErrInvalidParam
	}
	// 临时应用配置用于发送（不落库）
	tempConfig := map[string]any{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"username": cfg.Username,
		"password": cfg.Password,
		"from":     cfg.From,
	}
	code, err := s.verifySvc.GenerateCode()
	if err != nil {
		return "", err
	}
	identifier, err := s.verifySvc.GenerateIdentifier()
	if err != nil {
		return "", err
	}
	key := cache.GenerateEmailVerificationKey(to)
	if err := s.verifySvc.StoreCode(ctx, key, code); err != nil {
		return "", err
	}
	if err := s.emailServ.SendVerificationCodeWithConfig(tempConfig, to, code, identifier); err != nil {
		return "", err
	}
	return identifier, nil
}

// SendTestEmail 发送测试邮件，用于验证 SMTP 配置是否可用。
func (s *Service) SendTestEmail(ctx context.Context, to string) error {
	to = strings.TrimSpace(to)
	if s.emailServ == nil || !s.emailServ.IsConfigured() {
		return apperr.New(apperr.ErrInvalidParam.Code, "邮件服务未配置")
	}
	if to == "" {
		return apperr.ErrInvalidParam
	}
	if s.verifySvc == nil {
		return apperr.New(apperr.ErrInternal.Code, "验证码服务未初始化")
	}
	code, err := s.verifySvc.GenerateCode()
	if err != nil {
		return err
	}
	key := cache.GenerateEmailVerificationKey(to)
	if err := s.verifySvc.StoreCode(ctx, key, code); err != nil {
		return err
	}
	return s.emailServ.SendVerificationCode(to, code)
}

// VerifyEmailVerificationCode 校验邮箱验证码，通过后写入“已验证”标记（短期有效）。
func (s *Service) VerifyEmailVerificationCode(ctx context.Context, emailAddr, code string) error {
	emailAddr = strings.TrimSpace(emailAddr)
	code = strings.TrimSpace(code)
	if emailAddr == "" || code == "" {
		return apperr.ErrInvalidParam
	}
	if s.verifySvc == nil {
		return apperr.New(apperr.ErrInternal.Code, "验证码服务未初始化")
	}
	key := cache.GenerateEmailVerificationKey(emailAddr)
	stored, err := s.verifySvc.GetCode(ctx, key)
	if err != nil || stored != code {
		return apperr.ErrBadVerifyCode
	}
	_ = s.verifySvc.DeleteCode(ctx, key)
	// 设置“已验证”标记，供 /admin/settings 保存邮箱配置时校验
	markKey := cache.GenerateEmailVerifiedKey(emailAddr)
	// 10 分钟内允许保存一次邮箱配置
	return s.verifySvc.StoreCodeWithTTL(ctx, markKey, "1", 10*time.Minute)
}

func (s *Service) GetHotSearches(ctx context.Context) ([]string, error) {
	st, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if st.UIConfig == nil {
		return []string{}, nil
	}
	hotSearchStr, ok := st.UIConfig["hotSearchWords"].(string)
	if !ok || hotSearchStr == "" {
		return []string{}, nil
	}
	// 支持单引号格式的数组，转换为标准 JSON
	hotSearchStr = strings.ReplaceAll(hotSearchStr, "'", "\"")
	var hotSearches []string
	if err := json.Unmarshal([]byte(hotSearchStr), &hotSearches); err != nil {
		return []string{}, nil
	}
	return hotSearches, nil
}

func (s *Service) SaveMeiliSearchSyncTime(ctx context.Context, syncTime string) error {
	if err := s.repo.SaveMeiliSearchSyncTime(ctx, syncTime); err != nil {
		return err
	}
	if s.bizCache != nil {
		_ = s.bizCache.DeletePattern(ctx, "blog:setting:*")
	}
	return nil
}

// GetHotSearchMode 返回热门搜索模式："FAKE" 伪热门，"REAL" 真热门
func (s *Service) GetPrivacyPolicy(ctx context.Context) (string, error) {
	st, err := s.repo.Get(ctx)
	if err != nil {
		return "", err
	}
	if st.CoreConfig == nil {
		return "", nil
	}
	policy, ok := st.CoreConfig["privacyPolicy"].(string)
	if !ok || policy == "" {
		return "", nil
	}
	return policy, nil
}

func (s *Service) GetHotSearchMode(ctx context.Context) (string, error) {
	st, err := s.repo.Get(ctx)
	if err != nil {
		return "", err
	}
	if st.UIConfig == nil {
		return "FAKE", nil
	}
	// 支持布尔值（true=REAL, false=FAKE）、字符串（"REAL"/"FAKE"）和数字（0/1）三种格式
	if modeBool, ok := st.UIConfig["hotSearchMode"].(bool); ok {
		if modeBool {
			return "REAL", nil
		}
		return "FAKE", nil
	}
	if modeNum, ok := st.UIConfig["hotSearchMode"].(float64); ok {
		if modeNum != 0 {
			return "REAL", nil
		}
		return "FAKE", nil
	}
	mode, ok := st.UIConfig["hotSearchMode"].(string)
	if !ok || mode == "" {
		return "FAKE", nil
	}
	return mode, nil
}

func (s *Service) ExportSettings(ctx context.Context) (map[string]any, error) {
	st, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"coreConfig":    st.CoreConfig,
		"uiConfig":      st.UIConfig,
		"storageConfig": st.StorageConfig,
		"emailConfig":   st.EmailConfig,
	}, nil
}

func (s *Service) ImportSettings(ctx context.Context, data map[string]any, operator string) error {
	st, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}

	if v, ok := data["coreConfig"].(map[string]any); ok {
		st.CoreConfig = v
	}
	if v, ok := data["uiConfig"].(map[string]any); ok {
		st.UIConfig = v
	}
	if v, ok := data["storageConfig"].(map[string]any); ok {
		st.StorageConfig = v
	}
	if v, ok := data["emailConfig"].(map[string]any); ok {
		st.EmailConfig = v
	}

	if err := s.repo.Update(ctx, st, operator); err != nil {
		return err
	}

	if s.emailServ != nil && st.EmailConfig != nil {
		s.emailServ.UpdateConfig(st.EmailConfig)
	}
	if s.bizCache != nil {
		_ = s.bizCache.DeletePattern(ctx, "blog:setting:*")
	}
	return nil
}

func (s *Service) GetCoreConfig(ctx context.Context) (map[string]any, error) {
	cfg, err := s.repo.GetCoreConfig(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any(cfg), nil
}

func (s *Service) UpdateCoreConfig(ctx context.Context, body map[string]any, operator string) error {
	if err := s.repo.UpdateCoreConfig(ctx, domsetting.CoreConfig(body), operator); err != nil {
		return err
	}
	if s.bizCache != nil {
		_ = s.bizCache.DeletePattern(ctx, "blog:setting:*")
	}
	return nil
}

func (s *Service) GetSocialConfig(ctx context.Context) (map[string]any, error) {
	cfg, err := s.repo.GetSocialConfig(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any(cfg), nil
}

func (s *Service) UpdateSocialConfig(ctx context.Context, body map[string]any, operator string) error {
	if err := s.repo.UpdateSocialConfig(ctx, domsetting.SocialConfig(body), operator); err != nil {
		return err
	}
	if s.bizCache != nil {
		_ = s.bizCache.DeletePattern(ctx, "blog:setting:*")
	}
	return nil
}

func (s *Service) GetSearchConfig(ctx context.Context) (map[string]any, error) {
	cfg, err := s.repo.GetSearchConfig(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any(cfg), nil
}

func (s *Service) UpdateSearchConfig(ctx context.Context, body map[string]any, operator string) error {
	if err := s.repo.UpdateSearchConfig(ctx, domsetting.SearchConfig(body), operator); err != nil {
		return err
	}
	if s.bizCache != nil {
		_ = s.bizCache.DeletePattern(ctx, "blog:setting:*")
	}
	return nil
}

func (s *Service) GetPrivacyConfig(ctx context.Context) (map[string]any, error) {
	cfg, err := s.repo.GetPrivacyConfig(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any(cfg), nil
}

func (s *Service) UpdatePrivacyConfig(ctx context.Context, body map[string]any, operator string) error {
	if err := s.repo.UpdatePrivacyConfig(ctx, domsetting.PrivacyConfig(body), operator); err != nil {
		return err
	}
	if s.bizCache != nil {
		_ = s.bizCache.DeletePattern(ctx, "blog:setting:*")
	}
	return nil
}

func (s *Service) GetStorageConfig(ctx context.Context) (map[string]any, error) {
	cfg, err := s.repo.GetStorageConfig(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any(cfg), nil
}

func (s *Service) UpdateStorageConfig(ctx context.Context, body map[string]any, operator string) error {
	if err := s.repo.UpdateStorageConfig(ctx, domsetting.StorageConfig(body), operator); err != nil {
		return err
	}
	if s.bizCache != nil {
		_ = s.bizCache.DeletePattern(ctx, "blog:setting:*")
	}
	return nil
}

func (s *Service) GetEmailConfig(ctx context.Context) (map[string]any, error) {
	cfg, err := s.repo.GetEmailConfig(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any(cfg), nil
}

func (s *Service) UpdateEmailConfig(ctx context.Context, body map[string]any, operator string) error {
	newEmailCfg := body
	newUsername, _ := newEmailCfg["username"].(string)
	newUsername = strings.TrimSpace(newUsername)
	if newUsername != "" {
		needVerify := true
		oldCfg, err := s.repo.GetEmailConfig(ctx)
		if err == nil && len(oldCfg) > 0 {
			old := map[string]any(oldCfg)
			get := func(m map[string]any, k string) string {
				if v, ok := m[k].(string); ok {
					return strings.TrimSpace(v)
				}
				return ""
			}
			needVerify =
				get(old, "host") != get(newEmailCfg, "host") ||
					get(old, "port") != get(newEmailCfg, "port") ||
					get(old, "username") != get(newEmailCfg, "username") ||
					get(old, "password") != get(newEmailCfg, "password") ||
					get(old, "from") != get(newEmailCfg, "from")
		}
		if needVerify {
			if s.verifySvc == nil {
				return apperr.New(apperr.ErrInternal.Code, "验证码服务未初始化")
			}
			markKey := cache.GenerateEmailVerifiedKey(newUsername)
			mark, err := s.verifySvc.GetCode(ctx, markKey)
			if err != nil || mark != "1" {
				return apperr.ErrEmailNotVerified
			}
		}
	}

	if err := s.repo.UpdateEmailConfig(ctx, domsetting.EmailConfig(body), operator); err != nil {
		return err
	}

	if u, ok := newEmailCfg["username"].(string); ok && strings.TrimSpace(u) != "" && s.verifySvc != nil {
		_ = s.verifySvc.DeleteCode(ctx, cache.GenerateEmailVerifiedKey(strings.TrimSpace(u)))
	}

	if s.emailServ != nil {
		s.emailServ.UpdateConfig(body)
	}
	if s.bizCache != nil {
		_ = s.bizCache.DeletePattern(ctx, "blog:setting:*")
	}
	return nil
}
