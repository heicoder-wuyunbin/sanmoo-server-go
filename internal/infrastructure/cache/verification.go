package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// VerificationCodeExpiration 验证码过期时间
	VerificationCodeExpiration = 15 * time.Minute
	// VerificationCodeLength 验证码长度
	VerificationCodeLength = 6
)

// VerificationService 验证码服务
type VerificationService struct {
	client *redis.Client
}

// NewVerificationService 创建验证码服务实例
func NewVerificationService(client *redis.Client) *VerificationService {
	return &VerificationService{client: client}
}

// GenerateCode 生成验证码
func (s *VerificationService) GenerateCode() (string, error) {
	bytes := make([]byte, VerificationCodeLength/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// StoreCode 存储验证码到Redis
func (s *VerificationService) StoreCode(ctx context.Context, key string, code string) error {
	return s.client.Set(ctx, key, code, VerificationCodeExpiration).Err()
}

// StoreCodeWithTTL 存储验证码/标记到 Redis，并指定 TTL。
func (s *VerificationService) StoreCodeWithTTL(ctx context.Context, key string, code string, ttl time.Duration) error {
	return s.client.Set(ctx, key, code, ttl).Err()
}

// GetCode 从Redis获取验证码
func (s *VerificationService) GetCode(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

// DeleteCode 从Redis删除验证码
func (s *VerificationService) DeleteCode(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// GenerateLoginVerificationKey 生成登录验证码的Redis键
func GenerateLoginVerificationKey(userID uint64) string {
	return fmt.Sprintf("login_verification:%d", userID)
}

// GenerateEmailVerificationKey 生成邮箱验证码的Redis键
func GenerateEmailVerificationKey(email string) string {
	return fmt.Sprintf("email_verification:%s", email)
}

// GenerateEmailVerifiedKey 邮箱已验证标记（短期有效），用于“先验证后保存配置”的流程。
func GenerateEmailVerifiedKey(email string) string {
	return fmt.Sprintf("email_verified:%s", email)
}
