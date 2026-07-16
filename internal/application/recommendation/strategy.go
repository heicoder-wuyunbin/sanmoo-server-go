package recommendation

import (
	"context"

	domarticle "sanmoo-server-go/internal/domain/article"
)

// ArticleQuerier 查询已发布文章的最小接口。
type ArticleQuerier interface {
	ListArticles(ctx context.Context, q domarticle.ListQuery) ([]domarticle.Article, int64, error)
}

// GetPublishedArticles 返回已发布的最新文章（规则推荐）。
func GetPublishedArticles(ctx context.Context, repo ArticleQuerier, limit int) ([]domarticle.Article, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	isPublished := true
	articles, _, err := repo.ListArticles(ctx, domarticle.ListQuery{Page: 1, Size: limit, IsPublished: &isPublished})
	return articles, err
}