package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sanmoo-server-go/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 每分钟允许的请求次数
	LimitPerMinute int
	// 限流 Key 的前缀
	KeyPrefix string
	// 白名单 IP（不限流）
	WhitelistIPs []string
	// 白名单路径（不限流）
	WhitelistPaths []string
}

// 预定义限流策略
var (
	// 公开读接口限流配置（文章列表/详情/分类/标签等）
	PublicReadRateLimit = RateLimitConfig{
		LimitPerMinute: 60,
		KeyPrefix:      "ratelimit:read:",
		WhitelistPaths: []string{"/health", "/sitemap.xml", "/rss.xml", "/swagger", "/swagger/"},
	}

	// 公开写接口限流配置（点赞等）
	PublicWriteRateLimit = RateLimitConfig{
		LimitPerMinute: 10,
		KeyPrefix:      "ratelimit:write:",
		WhitelistPaths: []string{},
	}

	// 搜索接口限流配置
	SearchRateLimit = RateLimitConfig{
		LimitPerMinute: 20,
		KeyPrefix:      "ratelimit:search:",
		WhitelistPaths: []string{},
	}

	// 认证接口限流配置（登录/验证码等，防暴力破解）
	AuthRateLimit = RateLimitConfig{
		LimitPerMinute: 5,
		KeyPrefix:      "ratelimit:auth:",
		WhitelistPaths: []string{},
	}

	// 管理后台接口限流配置（JWT 身份限流，相对宽松）
	AdminRateLimit = RateLimitConfig{
		LimitPerMinute: 300,
		KeyPrefix:      "ratelimit:admin:",
		WhitelistPaths: []string{},
	}
)

// RateLimit 基于 Redis 的滑动窗口限流中间件
func RateLimit(redisClient *redis.Client, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 白名单路径直接跳过
		path := c.Request.URL.Path
		for _, wp := range config.WhitelistPaths {
			if path == wp || strings.HasPrefix(path, wp) {
				c.Next()
				return
			}
		}

		// 获取客户端 IP
		ip := c.ClientIP()

		// 白名单 IP 直接跳过（如内网 IP）
		for _, wip := range config.WhitelistIPs {
			if ip == wip || strings.HasPrefix(ip, wip) {
				c.Next()
				return
			}
		}

		// 构建限流 Key：ratelimit:read:192.168.1.1:/web/articles
		key := fmt.Sprintf("%s%s:%s", config.KeyPrefix, ip, path)

		// 使用 Redis INCR + EXPIRE 实现简单计数限流
		ctx := context.Background()
		val, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// Redis 错误不影响业务，记录警告后放行
			logger.Warnf("限流 Redis INCR 失败 key=%s: %v", key, err)
			c.Next()
			return
		}

		// 第一次请求时设置过期时间（1 分钟窗口）
		if val == 1 {
			redisClient.Expire(ctx, key, time.Minute)
		}

		// 获取剩余 TTL
		ttl, _ := redisClient.TTL(ctx, key).Result()
		retryAfter := int(ttl.Seconds())
		if retryAfter < 0 {
			retryAfter = 60
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.LimitPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(config.LimitPerMinute - int(val)))
		c.Header("X-RateLimit-Reset", strconv.Itoa(retryAfter))

		// 检查是否超限
		if int(val) > config.LimitPerMinute {
			logger.Warnf("限流触发: ip=%s path=%s count=%d limit=%d", ip, path, val, config.LimitPerMinute)
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":      false,
				"errorCode":    "RATE_LIMIT_EXCEEDED",
				"errorMessage": "请求过于频繁，请稍后再试",
				"data":         struct{}{},
				"timestamp":    time.Now().UnixMilli(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LikeRateLimit 点赞接口专项防刷：同一 IP 对同一文章 24 小时内只能点赞 1 次
func LikeRateLimit(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		articleID := c.Param("id")
		if articleID == "" {
			c.Next()
			return
		}

		// 构建防刷 Key：ratelimit:like:192.168.1.1:article:123
		key := fmt.Sprintf("ratelimit:like:%s:article:%s", ip, articleID)

		ctx := context.Background()
		// 使用 SETNX 判断是否已点赞（24 小时窗口）
		ok, err := redisClient.SetNX(ctx, key, "1", 24*time.Hour).Result()
		if err != nil {
			logger.Warnf("点赞防刷 Redis SETNX 失败 key=%s: %v", key, err)
			c.Next()
			return
		}

		if !ok {
			// 已存在，说明 24 小时内已点赞过
			logger.Warnf("点赞防刷触发: ip=%s articleID=%s", ip, articleID)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":      false,
				"errorCode":    "LIKE_LIMIT_EXCEEDED",
				"errorMessage": "24 小时内已点赞过该文章",
				"data":         struct{}{},
				"timestamp":    time.Now().UnixMilli(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}