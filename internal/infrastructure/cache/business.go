package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"sanmoo-server-go/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// 缓存 Key 前缀
const (
	KeyPrefix = "blog:"

	KeyArticleList        = "blog:article:list:%s"
	KeyArticleDetail      = "blog:article:%d"
	KeyArticlePrev        = "blog:article:%d:prev"
	KeyArticleNext        = "blog:article:%d:next"
	KeyArticleRelated     = "blog:article:%d:related"
	KeyArticleListPattern = "blog:article:list:*"
	KeyCategoryAll        = "blog:category:all"
	KeyTagAll             = "blog:tag:all"
	KeyDashboardStats     = "blog:dashboard:stats"
	KeySettingPub         = "blog:setting:pub"
	KeySetting            = "blog:setting:%s"
	KeyArchiveAll         = "blog:archive:all"
	KeyHotSearches        = "blog:hot:searches"
	KeySitemapArticles    = "blog:sitemap:articles"

	// 默认缓存 TTL（分钟），加随机偏移防止缓存雪崩
	DefaultTTLMin = 5
	ShortTTLMin   = 3
	LongTTLMin    = 10
)

// BusinessCache 业务缓存服务，封装对 Redis 的读写操作。
type BusinessCache struct {
	client    *redis.Client
	hitCount  int64
	missCount int64
}

// NewBusinessCache 创建业务缓存实例。
func NewBusinessCache(client *redis.Client) *BusinessCache {
	return &BusinessCache{client: client}
}

// Client 返回底层 Redis 客户端（供高级操作使用）。
func (c *BusinessCache) Client() *redis.Client {
	return c.client
}

// ttlWithJitter 生成带随机偏移的 TTL，防止缓存雪崩。
func ttlWithJitter(baseMin int) time.Duration {
	jitter := rand.Intn(60) // 0~60 秒随机偏移
	return time.Duration(baseMin)*time.Minute + time.Duration(jitter)*time.Second
}

// hashKey 对字符串做 MD5 哈希，用于生成缓存 key 的一部分。
func hashKey(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// BuildArticleListKey 构建文章列表缓存 key。
func BuildArticleListKey(page, size int, keyword string, categoryID, tagID uint64, isPublished string) string {
	raw := fmt.Sprintf("%d_%d_%s_%d_%d_%s", page, size, keyword, categoryID, tagID, isPublished)
	return fmt.Sprintf(KeyArticleList, hashKey(raw))
}

// Get 从缓存读取并反序列化到目标对象。
func (c *BusinessCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		c.missCount++
		return false, nil
	}
	if err != nil {
		logger.Warnf("Redis GET 失败 key=%s: %v", key, err)
		return false, err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		logger.Warnf("Redis 反序列化失败 key=%s: %v", key, err)
		return false, err
	}
	c.hitCount++
	return true, nil
}

// GetOrSet 尝试从缓存读取，未命中则调用 loader 回源并写入缓存。
// 缓存失败不会中断业务流程，仅记录警告日志。
func (c *BusinessCache) GetOrSet(
	ctx context.Context,
	key string,
	dest interface{},
	loader func() (interface{}, error),
	ttlMin int,
) error {
	if c == nil {
		result, err := loader()
		if err != nil {
			return err
		}
		return copyResult(dest, result)
	}

	hit, err := c.Get(ctx, key, dest)
	if err == nil && hit {
		return nil
	}

	result, err := loader()
	if err != nil {
		return err
	}

	if err := copyResult(dest, result); err != nil {
		return err
	}

	if writeErr := c.SetWithTTL(ctx, key, result, ttlMin); writeErr != nil {
		logger.Warnf("Redis SET 失败 key=%s: %v", key, writeErr)
	}

	return nil
}

// copyResult 通过 JSON 序列化/反序列化将 src 拷贝到 dest。
func copyResult(dest interface{}, src interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set 将对象序列化后写入缓存，使用默认 TTL。
func (c *BusinessCache) Set(ctx context.Context, key string, value interface{}) error {
	return c.SetWithTTL(ctx, key, value, DefaultTTLMin)
}

// SetWithTTL 将对象序列化后写入缓存，指定 TTL（分钟）。
func (c *BusinessCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttlMin int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttlWithJitter(ttlMin)).Err()
}

// Delete 删除单个缓存 key。
func (c *BusinessCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// DeletePattern 按模式删除缓存（如 blog:article:*）。
func (c *BusinessCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 500).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		// 分批删除，避免一次性删除大量 key 阻塞 Redis
		if len(keys) >= 100 {
			_ = c.client.Del(ctx, keys...)
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

// FlushBusiness 清空所有业务缓存（blog:* 前缀）。
func (c *BusinessCache) FlushBusiness(ctx context.Context) error {
	pattern := KeyPrefix + "*"
	iter := c.client.Scan(ctx, 0, pattern, 500).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			_ = c.client.Del(ctx, keys...)
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

// GetStats 获取缓存统计信息。
func (c *BusinessCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.client.Info(ctx, "memory", "keyspace").Result()
	if err != nil {
		return nil, err
	}

	// 统计 blog:* 前缀的 key 数量
	pattern := KeyPrefix + "*"
	iter := c.client.Scan(ctx, 0, pattern, 500).Iterator()
	var totalKeys int
	prefixCounts := make(map[string]int)
	for iter.Next(ctx) {
		totalKeys++
		key := iter.Val()
		// 按前缀分组统计
		prefix := extractPrefix(key)
		prefixCounts[prefix]++
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	// 解析内存信息
	memoryUsed := "unknown"
	for _, line := range strings.Split(info, "\r\n") {
		if strings.HasPrefix(line, "used_memory_human:") {
			memoryUsed = strings.TrimPrefix(line, "used_memory_human:")
		}
	}

	// 计算缓存命中率
	total := c.hitCount + c.missCount
	var hitRate float64
	if total > 0 {
		hitRate = float64(c.hitCount) / float64(total) * 100
	}

	return map[string]interface{}{
		"totalKeys":    totalKeys,
		"prefixCounts": prefixCounts,
		"memoryUsed":   memoryUsed,
		"hitRate":      fmt.Sprintf("%.2f", hitRate),
		"hitCount":     c.hitCount,
		"missCount":    c.missCount,
	}, nil
}

// extractPrefix 从 key 中提取前缀分组（如 blog:article:list → article:list）。
func extractPrefix(key string) string {
	withoutPrefix := strings.TrimPrefix(key, KeyPrefix)
	parts := strings.SplitN(withoutPrefix, ":", 2)
	if len(parts) >= 2 {
		// 对带 hash 的 key 只保留类型前缀
		return parts[0] + ":" + parts[1]
	}
	return parts[0]
}
