package recommendation

import (
	"context"

	domarticle "sanmoo-server-go/internal/domain/article"
)

const (
	StrategyRule     = "rule"
	StrategyWeighted = "weighted"
	StrategyCF       = "cf"
)

type Context struct {
	CurrentArticleID uint64
	Limit            int
}

type Strategy interface {
	Name() string
	Recommend(ctx context.Context, in Context) ([]domarticle.Article, error)
}

type ListArticleQuerier interface {
	ListArticle(ctx context.Context, q domarticle.ListQuery) ([]domarticle.Article, int64, error)
	FindByIDArticle(ctx context.Context, id uint64) (*domarticle.Article, error)
}
