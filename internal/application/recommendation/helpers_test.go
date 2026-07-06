package recommendation

import (
	"context"
	"testing"
	"time"

	domarticle "sanmoo-server-go/internal/domain/article"
)

type fakeListArticleRepo struct {
	items []domarticle.Article
}

func (f *fakeListArticleRepo) ListArticle(_ context.Context, _ domarticle.ListQuery) ([]domarticle.Article, int64, error) {
	return f.items, int64(len(f.items)), nil
}

func (f *fakeListArticleRepo) FindByIDArticle(_ context.Context, id uint64) (*domarticle.Article, error) {
	for i := range f.items {
		if f.items[i].ID == id {
			return &f.items[i], nil
		}
	}
	return nil, nil
}

func TestListPublishedArticles_DeduplicatesByID(t *testing.T) {
	repo := &fakeListArticleRepo{
		items: []domarticle.Article{
			{ID: 34, CategoryID: 1, Category: "Java"},
			{ID: 34, CategoryID: 4, Category: "多线程"},
			{ID: 33, CategoryID: 1, Category: "Java"},
		},
	}

	got, err := listPublishedArticles(context.Background(), repo, 10)
	if err != nil {
		t.Fatalf("listPublishedArticles returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 unique articles, got %d", len(got))
	}
	if got[0].ID != 34 || got[1].ID != 33 {
		t.Fatalf("unexpected article ids: %+v", []uint64{got[0].ID, got[1].ID})
	}
}

func TestTopScored_DeduplicatesAndRespectsLimit(t *testing.T) {
	now := time.Now()
	items := []scoredArticle{
		{article: domarticle.Article{ID: 1, CreateTime: now}, score: 100},
		{article: domarticle.Article{ID: 1, CreateTime: now.Add(-time.Minute)}, score: 99},
		{article: domarticle.Article{ID: 2, CreateTime: now.Add(-2 * time.Minute)}, score: 98},
	}

	got := topScored(items, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 articles, got %d", len(got))
	}
	if got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("unexpected ordering/result ids: %+v", []uint64{got[0].ID, got[1].ID})
	}
}
