package tag

import "context"

type ListQuery struct {
	Page    int
	Size    int
	Keyword string
}

type Repository interface {
	ListTags(ctx context.Context, q ListQuery) ([]Tag, int64, error)
	ListAllWithCountTag(ctx context.Context) ([]Tag, error)
	FindByIDTag(ctx context.Context, id uint64) (*Tag, error)
	ListTagsByIDs(ctx context.Context, ids []uint64) ([]Tag, error)
	CreateTag(ctx context.Context, t *Tag) (uint64, error)
	UpdateTag(ctx context.Context, t *Tag) error
	DeleteTag(ctx context.Context, id uint64) error
	BatchDeleteTag(ctx context.Context, ids []uint64) error
}
