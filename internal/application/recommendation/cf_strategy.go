package recommendation

import (
	"context"

	domarticle "sanmoo-server-go/internal/domain/article"
)

type CFStrategy struct {
	repo ListArticleQuerier
}

func NewCFStrategy(repo ListArticleQuerier) *CFStrategy {
	return &CFStrategy{repo: repo}
}

func (s *CFStrategy) Name() string { return StrategyCF }

func (s *CFStrategy) Recommend(ctx context.Context, in Context) ([]domarticle.Article, error) {
	limit := normalizeLimit(in.Limit)
	candidates, err := listPublishedArticles(ctx, s.repo, 200)
	if err != nil {
		return nil, err
	}

	if in.CurrentArticleID == 0 {
		// 冷启动时，CF 退化为加权热点+新鲜度。
		weighted := NewWeightedStrategy(s.repo)
		return weighted.Recommend(ctx, in)
	}

	current, err := s.repo.FindByIDArticle(ctx, in.CurrentArticleID)
	if err != nil || current == nil {
		rule := NewRuleStrategy(s.repo)
		return rule.Recommend(ctx, in)
	}

	baseTags := tagSet(current.Tags)
	scored := make([]scoredArticle, 0, len(candidates))
	for _, a := range candidates {
		if a.ID == in.CurrentArticleID {
			continue
		}
		tags := tagSet(a.Tags)
		inter := overlapCount(baseTags, tags)
		if inter == 0 {
			continue
		}
		union := len(baseTags) + len(tags) - inter
		sim := float64(inter) / float64(union)
		score := sim*100 + float64(a.ReadNum)*0.2
		scored = append(scored, scoredArticle{article: a, score: score})
	}
	if len(scored) == 0 {
		weighted := NewWeightedStrategy(s.repo)
		return weighted.Recommend(ctx, in)
	}
	return topScored(scored, limit), nil
}
