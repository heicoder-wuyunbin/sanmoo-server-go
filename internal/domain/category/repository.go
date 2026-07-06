package category

import "context"

type Repository interface {
	ListAllWithCountCategory(ctx context.Context) ([]Category, error)
	FindByIDCategory(ctx context.Context, id uint64) (*Category, error)
	ListCategoriesByIDs(ctx context.Context, ids []uint64) ([]Category, error)
	CreateCategory(ctx context.Context, c *Category) (uint64, error)
	UpdateCategory(ctx context.Context, c *Category) error
	DeleteCategory(ctx context.Context, id uint64) error
	BatchDeleteCategory(ctx context.Context, ids []uint64) error
}
