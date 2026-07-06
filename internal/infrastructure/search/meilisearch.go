package search

import (
	"context"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

type MeiliSearchClient struct {
	client meilisearch.ServiceManager
	index  string
}

func NewMeiliSearchClient(host, apiKey, index string) *MeiliSearchClient {
	var opts []meilisearch.Option
	if apiKey != "" {
		opts = append(opts, meilisearch.WithAPIKey(apiKey))
	}
	client := meilisearch.New(host, opts...)
	return &MeiliSearchClient{
		client: client,
		index:  index,
	}
}

func (m *MeiliSearchClient) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if query == "" {
		return []SearchResult{}, nil
	}

	idx := m.client.Index(m.index)
	searchRes, err := idx.Search(query, &meilisearch.SearchRequest{
		Limit: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	var results []SearchResult
	for _, hit := range searchRes.Hits {
		var result map[string]interface{}
		if err := hit.Decode(&result); err != nil {
			continue
		}
		results = append(results, SearchResult{
			ID:    int64(result["id"].(float64)),
			Title: result["title"].(string),
			Desc:  getString(result, "description"),
		})
	}
	return results, nil
}

func (m *MeiliSearchClient) AddDocuments(ctx context.Context, docs []interface{}) error {
	idx := m.client.Index(m.index)
	_, err := idx.AddDocuments(docs, nil)
	if err != nil {
		return fmt.Errorf("meilisearch add documents failed: %w", err)
	}
	return nil
}

func (m *MeiliSearchClient) CreateIndexIfNotExists(ctx context.Context) error {
	idx := m.client.Index(m.index)
	_, err := idx.GetSettings()
	if err == nil {
		return nil
	}

	_, err = m.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        m.index,
		PrimaryKey: "id",
	})
	if err != nil {
		return fmt.Errorf("meilisearch create index failed: %w", err)
	}

	_, err = idx.UpdateSettings(&meilisearch.Settings{
		SearchableAttributes: []string{"title", "description", "content"},
		DisplayedAttributes:  []string{"id", "title", "description"},
	})
	if err != nil {
		return fmt.Errorf("meilisearch update settings failed: %w", err)
	}

	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

type SearchResult struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
}
