package article

import "context"

type ListQuery struct {
	Page        int
	Size        int
	Keyword     string
	CategoryID  uint64
	TagID       uint64
	IsPublished *bool
}

type ArchiveItem struct {
	Month string    `json:"month"`
	Items []Article `json:"items"`
}

type PVPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type Repository interface {
	ListArticles(ctx context.Context, q ListQuery) ([]Article, int64, error)
	FindByIDArticle(ctx context.Context, id uint64) (*Article, error)
	FindPrev(ctx context.Context, id uint64) (*Article, error)
	FindNext(ctx context.Context, id uint64) (*Article, error)
	CreateArticle(ctx context.Context, a *Article) (uint64, error)
	UpdateArticle(ctx context.Context, a *Article) error
	UpdateArticleStatus(ctx context.Context, id uint64, isPublished, isTop *bool) error
	BatchUpdateArticleStatus(ctx context.Context, ids []uint64, isPublished, isTop *bool) error
	DeleteArticle(ctx context.Context, id uint64) error
	BatchDeleteArticle(ctx context.Context, ids []uint64) error
	ArchiveList(ctx context.Context) ([]ArchiveItem, error)
	IncreaseReadAndPV(ctx context.Context, articleID uint64) error
	IncreaseShareCount(ctx context.Context, articleID uint64) error
	LikeArticle(ctx context.Context, articleID uint64) (int, error)
	PVStatistics(ctx context.Context, days int) ([]PVPoint, error)
	ListArticlesByIDs(ctx context.Context, ids []uint64) ([]Article, error)
	RecordSearchHistory(ctx context.Context, keyword string) error
	GetHotSearchKeywords(ctx context.Context, limit int) ([]string, error)
	RandomArticle(ctx context.Context, excludeID uint64) (*Article, error)
}
