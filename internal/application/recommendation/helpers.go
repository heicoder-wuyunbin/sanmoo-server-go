package recommendation

import (
	"context"
	"sort"
	"time"

	domarticle "sanmoo-server-go/internal/domain/article"
)

type scoredArticle struct {
	article domarticle.Article
	score   float64
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func listPublishedArticles(ctx context.Context, repo ListArticleQuerier, size int) ([]domarticle.Article, error) {
	items, _, err := repo.ListArticle(ctx, domarticle.ListQuery{Page: 1, Size: size, IsPublished: boolPtr(true)})
	if err != nil {
		return nil, err
	}
	return dedupeArticlesByID(items), nil
}

func tagSet(tags []domarticle.TagRef) map[uint64]struct{} {
	m := make(map[uint64]struct{}, len(tags))
	for _, t := range tags {
		m[t.ID] = struct{}{}
	}
	return m
}

func overlapCount(a, b map[uint64]struct{}) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := 0
	for id := range a {
		if _, ok := b[id]; ok {
			n++
		}
	}
	return n
}

func topScored(items []scoredArticle, limit int) []domarticle.Article {
	sort.Slice(items, func(i, j int) bool {
		if items[i].score == items[j].score {
			return items[i].article.CreateTime.After(items[j].article.CreateTime)
		}
		return items[i].score > items[j].score
	})
	out := make([]domarticle.Article, 0, min(limit, len(items)))
	seen := make(map[uint64]struct{}, len(items))
	for _, it := range items {
		if len(out) >= limit {
			break
		}
		if _, ok := seen[it.article.ID]; ok {
			continue
		}
		seen[it.article.ID] = struct{}{}
		out = append(out, it.article)
	}
	return out
}

func dedupeArticlesByID(items []domarticle.Article) []domarticle.Article {
	if len(items) <= 1 {
		return items
	}
	out := make([]domarticle.Article, 0, len(items))
	seen := make(map[uint64]struct{}, len(items))
	for _, a := range items {
		if _, ok := seen[a.ID]; ok {
			continue
		}
		seen[a.ID] = struct{}{}
		out = append(out, a)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func daysSince(t time.Time) float64 {
	h := time.Since(t).Hours()
	if h < 0 {
		return 0
	}
	return h / 24
}

func boolPtr(v bool) *bool {
	return &v
}
