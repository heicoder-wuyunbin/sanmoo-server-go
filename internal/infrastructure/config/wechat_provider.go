package config

import (
	"context"
	"encoding/json"
	"fmt"

	"sanmoo-server-go/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

const wechatConfigCacheKey = "blog:wechat:config"

// WechatConfigProvider 动态微信配置提供者，支持 Redis 缓存 + DB 回退。
// 替代原有的配置文件/环境变量读取方式，支持后台动态切换开发/生产环境。
type WechatConfigProvider struct {
	redisClient *redis.Client
	loader      func(ctx context.Context) (*WechatMPConfig, error)
}

// NewWechatConfigProvider 创建微信配置提供者。
// loader 是 DB 回退函数，用于缓存未命中时从数据库加载配置。
func NewWechatConfigProvider(redisClient *redis.Client, loader func(ctx context.Context) (*WechatMPConfig, error)) *WechatConfigProvider {
	return &WechatConfigProvider{
		redisClient: redisClient,
		loader:      loader,
	}
}

// Get 获取当前环境对应的微信小程序配置。
// 优先从 Redis 缓存读取，缓存未命中时回退到数据库。
func (p *WechatConfigProvider) Get(ctx context.Context) (*WechatMPConfig, error) {
	// 1. 尝试从 Redis 缓存读取
	if p.redisClient != nil {
		val, err := p.redisClient.Get(ctx, wechatConfigCacheKey).Result()
		if err == nil {
			var cfg WechatMPConfig
			if err := json.Unmarshal([]byte(val), &cfg); err == nil {
				return &cfg, nil
			}
			logger.Warnf("微信配置缓存反序列化失败: %v", err)
		} else if err != redis.Nil {
			logger.Warnf("Redis 读取微信配置失败: %v", err)
		}
	}

	// 2. 缓存未命中，回退到数据库
	if p.loader == nil {
		return nil, fmt.Errorf("微信配置加载器未设置")
	}
	cfg, err := p.loader(ctx)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	if p.redisClient != nil {
		data, _ := json.Marshal(cfg)
		_ = p.redisClient.Set(ctx, wechatConfigCacheKey, data, 0).Err() // 永不过期，由更新时主动清除
	}

	return cfg, nil
}

// InvalidateCache 清除微信配置缓存（配置更新后调用）。
func (p *WechatConfigProvider) InvalidateCache(ctx context.Context) {
	if p.redisClient != nil {
		_ = p.redisClient.Del(ctx, wechatConfigCacheKey).Err()
	}
}

// LoadFromDB 从数据库加载微信配置，判断当前环境并返回对应的 AppID/Secret。
func LoadWechatConfigFromDB(ctx context.Context, wechatConfig map[string]any) *WechatMPConfig {
	envMode := int64(0)
	if v, ok := wechatConfig["wxEnvMode"]; ok {
		switch val := v.(type) {
		case float64:
			envMode = int64(val)
		case int64:
			envMode = val
		}
	}

	var appID, secret string
	if envMode == 1 {
		// 生产环境
		appID, _ = wechatConfig["wxProdAppId"].(string)
		secret, _ = wechatConfig["wxProdAppSecret"].(string)
	} else {
		// 开发环境（默认）
		appID, _ = wechatConfig["wxDevAppId"].(string)
		secret, _ = wechatConfig["wxDevAppSecret"].(string)
	}

	return &WechatMPConfig{
		AppID:  appID,
		Secret: secret,
	}
}