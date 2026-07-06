package cache

import (
	"context"
	"fmt"

	"sanmoo-server-go/internal/application/article"
	"sanmoo-server-go/internal/application/category"
	"sanmoo-server-go/internal/application/dashboard"
	"sanmoo-server-go/internal/application/setting"
	"sanmoo-server-go/internal/application/tag"
	bizcache "sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/infrastructure/logger"
)

// Service 缓存管理服务，提供一键清空和预热功能。
type Service struct {
	bizCache     *bizcache.BusinessCache
	articleSvc   *article.Service
	categorySvc  *category.Service
	tagSvc       *tag.Service
	dashboardSvc *dashboard.Service
	settingSvc   *setting.Service
}

// NewService 创建缓存管理服务实例。
func NewService(
	bizCache *bizcache.BusinessCache,
	articleSvc *article.Service,
	categorySvc *category.Service,
	tagSvc *tag.Service,
	dashboardSvc *dashboard.Service,
	settingSvc *setting.Service,
) *Service {
	return &Service{
		bizCache:     bizCache,
		articleSvc:   articleSvc,
		categorySvc:  categorySvc,
		tagSvc:       tagSvc,
		dashboardSvc: dashboardSvc,
		settingSvc:   settingSvc,
	}
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

// WarmupAll 一键缓存所有文章、分类、标签、设置、归档等数据。
func (s *Service) WarmupAll(ctx context.Context) (*WarmupResult, error) {
	result := &WarmupResult{}

	// 预热分类列表
	if s.categorySvc != nil {
		_, err := s.categorySvc.ListCategories(ctx)
		if err != nil {
			logger.Warnf("预热分类缓存失败: %v", err)
		} else {
			result.Categories = 1
		}
	}

	// 预热标签列表
	if s.tagSvc != nil {
		_, err := s.tagSvc.ListAllTags(ctx)
		if err != nil {
			logger.Warnf("预热标签缓存失败: %v", err)
		} else {
			result.Tags = 1
		}
	}

	// 预热文章列表（第一页，已发布）
	if s.articleSvc != nil {
		published := 1
		_, err := s.articleSvc.ListArticles(ctx, 1, 10, "", 0, 0, &published)
		if err != nil {
			logger.Warnf("预热文章列表缓存失败: %v", err)
		} else {
			result.Articles = 10
		}
	}

	// 预热归档
	if s.articleSvc != nil {
		_, err := s.articleSvc.Archives(ctx)
		if err != nil {
			logger.Warnf("预热归档缓存失败: %v", err)
		} else {
			result.Archives = true
		}
	}

	// 预热热门搜索
	if s.articleSvc != nil {
		_, err := s.articleSvc.GetRealHotSearches(ctx, 10)
		if err != nil {
			logger.Warnf("预热热门搜索缓存失败: %v", err)
		} else {
			result.HotSearches = true
		}
	}

	// 预热仪表盘
	if s.dashboardSvc != nil {
		_, err := s.dashboardSvc.Dashboard(ctx)
		if err != nil {
			logger.Warnf("预热仪表盘缓存失败: %v", err)
		} else {
			result.Dashboard = true
		}
	}

	// 预热站点设置
	if s.settingSvc != nil {
		_, err := s.settingSvc.GetSettings(ctx)
		if err != nil {
			logger.Warnf("预热设置缓存失败: %v", err)
		} else {
			result.Settings = true
		}
	}

	result.Message = "缓存预热完成"
	logger.Infof("缓存预热完成: categories=%v, tags=%v, articles=%d, dashboard=%v, settings=%v, archives=%v, hotSearches=%v",
		result.Categories, result.Tags, result.Articles, result.Dashboard, result.Settings, result.Archives, result.HotSearches)

	return result, nil
}

// Stats 获取缓存统计信息。
func (s *Service) Stats(ctx context.Context) (map[string]interface{}, error) {
	if s.bizCache == nil {
		return nil, fmt.Errorf("缓存服务未初始化")
	}
	return s.bizCache.GetStats(ctx)
}