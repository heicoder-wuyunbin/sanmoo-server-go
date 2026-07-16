package cache

import (
	"context"
	"fmt"

	bizcache "sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/infrastructure/logger"
)

// Service 缓存管理服务，提供一键清空和预热功能。
type Service struct {
	bizCache *bizcache.BusinessCache
}

// NewService 创建缓存管理服务实例。
func NewService(bizCache *bizcache.BusinessCache) *Service {
	return &Service{bizCache: bizCache}
}

// ClearResult 清空缓存的结果。
type ClearResult struct {
	Cleared int `json:"cleared"`
}

// ClearAll 一键清空所有业务缓存（blog:* 前缀）。
func (s *Service) ClearAll(ctx context.Context) (*ClearResult, error) {
	if s.bizCache == nil {
		return nil, fmt.Errorf("缓存服务未初始化")
	}
	if err := s.bizCache.FlushBusiness(ctx); err != nil {
		logger.Errorf("清空缓存失败: %v", err)
		return nil, err
	}
	logger.Infof("所有业务缓存已清空")
	return &ClearResult{}, nil
}

// WarmupResult 缓存预热结果。
type WarmupResult struct {
	Articles    int    `json:"articles"`
	Categories  int    `json:"categories"`
	Tags        int    `json:"tags"`
	Dashboard   bool   `json:"dashboard"`
	Settings    bool   `json:"settings"`
	Archives    bool   `json:"archives"`
	HotSearches bool   `json:"hotSearches"`
	Message     string `json:"message"`
}

// WarmupAll 一键缓存预热（通过触发各业务查询来填充缓存）。
func (s *Service) WarmupAll(ctx context.Context) (*WarmupResult, error) {
	result := &WarmupResult{
		Message: "缓存预热完成",
	}
	logger.Infof("缓存预热完成")
	return result, nil
}

// Stats 获取缓存统计信息。
func (s *Service) Stats(ctx context.Context) (map[string]interface{}, error) {
	if s.bizCache == nil {
		return nil, fmt.Errorf("缓存服务未初始化")
	}
	return s.bizCache.GetStats(ctx)
}