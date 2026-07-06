package dashboard

import (
	"context"

	domdashboard "sanmoo-server-go/internal/domain/dashboard"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/shared/pagination"
)

type Service struct {
	repo  domdashboard.Repository
	cache *cache.BusinessCache
}

func NewService(repo domdashboard.Repository, cache *cache.BusinessCache) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) Dashboard(ctx context.Context) (*dto.DashboardResponse, error) {
	// 尝试从缓存读取
	if s.cache != nil {
		var cached dto.DashboardResponse
		if hit, _ := s.cache.Get(ctx, cache.KeyDashboardStats, &cached); hit {
			return &cached, nil
		}
	}

	d, err := s.repo.Statistics(ctx)
	if err != nil {
		return nil, err
	}
	result := &dto.DashboardResponse{Dashboard: d}

	// 写入缓存
	if s.cache != nil {
		_ = s.cache.SetWithTTL(ctx, cache.KeyDashboardStats, result, cache.ShortTTLMin)
	}

	return result, nil
}

func (s *Service) VisitorRecords(ctx context.Context, page, size int, keyword string) (*pagination.PageData, error) {
	list, total, err := s.repo.VisitorList(ctx, page, size, keyword)
	if err != nil {
		return nil, err
	}
	return pagination.NewPageData(list, total, page, size), nil
}

func (s *Service) ErrorLogRecords(ctx context.Context, page, size int, keyword string) (*pagination.PageData, error) {
	list, total, err := s.repo.ErrorLogList(ctx, page, size, keyword)
	if err != nil {
		return nil, err
	}
	return pagination.NewPageData(list, total, page, size), nil
}

func (s *Service) DeleteVisitorRecord(ctx context.Context, id uint64) error {
	return s.repo.DeleteVisitorRecord(ctx, id)
}

func (s *Service) BatchDeleteVisitorRecords(ctx context.Context, ids []uint64) error {
	return s.repo.BatchDeleteVisitorRecords(ctx, ids)
}

func (s *Service) ClearAllVisitorRecords(ctx context.Context) error {
	return s.repo.ClearAllVisitorRecords(ctx)
}

func (s *Service) DeleteErrorLog(ctx context.Context, id uint64) error {
	return s.repo.DeleteErrorLog(ctx, id)
}

func (s *Service) BatchDeleteErrorLogs(ctx context.Context, ids []uint64) error {
	return s.repo.BatchDeleteErrorLogs(ctx, ids)
}

func (s *Service) ClearAllErrorLogs(ctx context.Context) error {
	return s.repo.ClearAllErrorLogs(ctx)
}

func (s *Service) ImportErrorLogs(ctx context.Context, logs []domdashboard.ErrorLogRecord) (int64, error) {
	return s.repo.ImportErrorLogs(ctx, logs)
}

func (s *Service) ExportErrorLogs(ctx context.Context) ([]domdashboard.ErrorLogRecord, error) {
	return s.repo.ExportErrorLogs(ctx)
}

func (s *Service) PV(ctx context.Context, days int) (any, error) {
	return s.repo.PVStatistics(ctx, days)
}

func (s *Service) TagStatistics(ctx context.Context) ([]domdashboard.NameValue, error) {
	return s.repo.TagStatistics(ctx)
}

func (s *Service) CategoryStatistics(ctx context.Context) ([]domdashboard.NameValue, error) {
	return s.repo.CategoryStatistics(ctx)
}

func (s *Service) ArticlePublishHeatmap(ctx context.Context) ([]domdashboard.DateCount, error) {
	return s.repo.ArticlePublishHeatmap(ctx)
}

func (s *Service) TopicStatistics(ctx context.Context) ([]domdashboard.NameValue, error) {
	return s.repo.TopicStatistics(ctx)
}

func (s *Service) MpUserGrowth(ctx context.Context, days int) ([]domdashboard.DateCount, error) {
	return s.repo.MpUserGrowth(ctx, days)
}

func (s *Service) ArticleReadStatistics(ctx context.Context, page, size int) (*pagination.PageData, error) {
	list, total, err := s.repo.ArticleReadStatistics(ctx, page, size)
	if err != nil {
		return nil, err
	}
	return pagination.NewPageData(list, total, page, size), nil
}

func (s *Service) CategoryReadStatistics(ctx context.Context) ([]domdashboard.CategoryReadStat, error) {
	return s.repo.CategoryReadStatistics(ctx)
}

func (s *Service) TagReadStatistics(ctx context.Context) ([]domdashboard.TagReadStat, error) {
	return s.repo.TagReadStatistics(ctx)
}

func (s *Service) ContentTrend(ctx context.Context, days int) ([]domdashboard.ContentTrend, error) {
	return s.repo.ContentTrend(ctx, days)
}
