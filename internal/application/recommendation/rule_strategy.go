package recommendation

import (
	"context"

	domarticle "sanmoo-server-go/internal/domain/article"
)

type RuleStrategy struct {
	repo ListArticleQuerier
}

func NewRuleStrategy(repo ListArticleQuerier) *RuleStrategy {
	return &RuleStrategy{repo: repo}
}

func (s *RuleStrategy) Name() string { return StrategyRule }

func (s *RuleStrategy) Recommend(ctx context.Context, in Context) ([]domarticle.Article, error) {
	limit := normalizeLimit(in.Limit)
	return listPublishedArticles(ctx, s.repo, limit)
}
