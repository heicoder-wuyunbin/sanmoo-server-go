package scheduler

import (
	"context"
	"fmt"
	"time"

	domarticle "sanmoo-server-go/internal/domain/article"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/infrastructure/logger"
)

// ScheduledPublishScheduler 文章定时发布调度器
type ScheduledPublishScheduler struct {
	repo   domarticle.Repository
	cache  *cache.BusinessCache
	ticker *time.Ticker
	stopCh chan struct{}
}

// NewScheduledPublishScheduler 创建定时发布调度器
func NewScheduledPublishScheduler(repo domarticle.Repository, cache *cache.BusinessCache) *ScheduledPublishScheduler {
	return &ScheduledPublishScheduler{
		repo:   repo,
		cache:  cache,
		stopCh: make(chan struct{}),
	}
}

// Start 启动调度器（每分钟检查一次）
func (s *ScheduledPublishScheduler) Start() {
	s.ticker = time.NewTicker(1 * time.Minute)
	logger.Infof("定时发布调度器启动，每分钟检查一次")

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkAndPublish()
			case <-s.stopCh:
				logger.Infof("定时发布调度器停止")
				return
			}
		}
	}()
}

// Stop 停止调度器
func (s *ScheduledPublishScheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopCh)
}

// checkAndPublish 检查并发布到时间的文章
func (s *ScheduledPublishScheduler) checkAndPublish() {
	ctx := context.Background()
	now := time.Now()

	logger.Debugf("定时发布调度器检查: %s", now.Format("2006-01-02 15:04:05"))

	// 查询需要发布的文章：is_published=0 且 publish_time <= NOW() 且 publish_time IS NOT NULL
	articles, err := s.repo.FindScheduledArticles(ctx, now)
	if err != nil {
		logger.Warnf("查询待发布文章失败: %v", err)
		return
	}

	if len(articles) == 0 {
		return
	}

	logger.Infof("发现 %d 篇待发布文章", len(articles))

	for _, article := range articles {
		// 发布文章：设置 is_published=1，清空 publish_time
		err := s.repo.PublishScheduledArticle(ctx, article.ID)
		if err != nil {
			logger.Warnf("发布文章 %d 失败: %v", article.ID, err)
			continue
		}

		logger.Infof("文章 %d (%s) 已自动发布", article.ID, article.Title)

		// 清理相关缓存
		if s.cache != nil {
			// 清理文章列表缓存
			_ = s.cache.DeletePattern(ctx, cache.KeyArticleListPattern)
			// 清理文章详情缓存
			detailKey := fmt.Sprintf(cache.KeyArticleDetail, article.ID)
			_ = s.cache.Delete(ctx, detailKey)
		}
	}
}