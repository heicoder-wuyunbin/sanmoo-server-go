package cache

import (
	"context"
	"time"

	"sanmoo-server-go/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

func NewRedis(addr string) *redis.Client {
	logger.Infof("正在初始化Redis客户端，地址: %s", addr)
	client := redis.NewClient(&redis.Options{Addr: addr, ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warnf("Redis连接测试失败: %v，将在需要时尝试重连", err)
	} else {
		logger.Infof("Redis连接测试成功")
	}

	return client
}

func Health(ctx context.Context, c *redis.Client) error {
	logger.Debugf("正在检查Redis健康状态")
	err := c.Ping(ctx).Err()
	if err != nil {
		logger.Warnf("Redis健康检查失败: %v", err)
	} else {
		logger.Debugf("Redis健康检查成功")
	}
	return err
}
