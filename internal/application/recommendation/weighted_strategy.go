package recommendation

import (
	"context"

	domarticle "sanmoo-server-go/internal/domain/article"
)

type WeightedStrategy struct {
	repo ListArticleQuerier
}

func NewWeightedStrategy(repo ListArticleQuerier) *WeightedStrategy {
	return &WeightedStrategy{repo: repo}
}

func (s *WeightedStrategy) Name() string { return StrategyWeighted }

func (s *WeightedStrategy) Recommend(ctx context.Context, in Context) ([]domarticle.Article, error) {
	limit := normalizeLimit(in.Limit)
	candidates, err := listPublishedArticles(ctx, s.repo, 200)
	if err != nil {
		return nil, err
	}

	var current *domarticle.Article
	if in.CurrentArticleID > 0 {
		current, _ = s.repo.FindByIDArticle(ctx, in.CurrentArticleID)
	}
	baseTags := map[uint64]struct{}{}
	if current != nil {
		baseTags = tagSet(current.Tags)
	}

	scored := make([]scoredArticle, 0, len(candidates))
	for _, a := range candidates {
		if in.CurrentArticleID > 0 && a.ID == in.CurrentArticleID {
			continue
		}
		freshness := 1.0 / (1.0 + daysSince(a.CreateTime)/7.0)
		hot := float64(a.ReadNum)
		overlap := float64(overlapCount(baseTags, tagSet(a.Tags)))
		if len(baseTags) == 0 {
			overlap = 0
		}
		score := hot*0.6 + freshness*30 + overlap*8
		scored = append(scored, scoredArticle{article: a, score: score})
	}
	return topScored(scored, limit), nil
}
