package dashboard

import (
	"context"
	"sanmoo-server-go/internal/domain/article"
)

type Repository interface {
	Statistics(ctx context.Context) (*Statistics, error)
	VisitorList(ctx context.Context, page, size int, keyword string) ([]VisitorRecord, int64, error)
	ErrorLogList(ctx context.Context, page, size int, keyword string) ([]ErrorLogRecord, int64, error)
	DeleteVisitorRecord(ctx context.Context, id uint64) error
	BatchDeleteVisitorRecords(ctx context.Context, ids []uint64) error
	ClearAllVisitorRecords(ctx context.Context) error
	DeleteErrorLog(ctx context.Context, id uint64) error
	BatchDeleteErrorLogs(ctx context.Context, ids []uint64) error
	ClearAllErrorLogs(ctx context.Context) error
	ImportErrorLogs(ctx context.Context, logs []ErrorLogRecord) (int64, error)
	ExportErrorLogs(ctx context.Context) ([]ErrorLogRecord, error)
	TagStatistics(ctx context.Context) ([]NameValue, error)
	CategoryStatistics(ctx context.Context) ([]NameValue, error)
	ArticlePublishHeatmap(ctx context.Context) ([]DateCount, error)
	PVStatistics(ctx context.Context, days int) ([]article.PVPoint, error)
	TopicStatistics(ctx context.Context) ([]NameValue, error)
	MpUserGrowth(ctx context.Context, days int) ([]DateCount, error)
	ArticleReadStatistics(ctx context.Context, page, size int) ([]ArticleReadStat, int64, error)
	CategoryReadStatistics(ctx context.Context) ([]CategoryReadStat, error)
	TagReadStatistics(ctx context.Context) ([]TagReadStat, error)
	ContentTrend(ctx context.Context, days int) ([]ContentTrend, error)
}
