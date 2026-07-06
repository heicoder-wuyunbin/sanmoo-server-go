package permission

import "context"

type ListQuery struct {
	Page     int
	Size     int
	Keyword  string
	Module   string
	Type     string
}

type Repository interface {
	ListPermissions(ctx context.Context, q ListQuery) ([]Permission, int64, error)
	ListAllPermissions(ctx context.Context) ([]Permission, error)
	FindByIDPermission(ctx context.Context, id uint64) (*Permission, error)
	FindByPermKey(ctx context.Context, permKey string) (*Permission, error)
	CreatePermission(ctx context.Context, p *Permission) (uint64, error)
	UpdatePermission(ctx context.Context, p *Permission) error
	DeletePermission(ctx context.Context, id uint64) error
	GetPermKeysByRoleIDs(ctx context.Context, roleIDs []uint64) ([]string, error)
}
