package role

import "context"

type ListQuery struct {
	Page    int
	Size    int
	Keyword string
}

type Repository interface {
	ListRoles(ctx context.Context, q ListQuery) ([]Role, int64, error)
	ListAllRoles(ctx context.Context) ([]Role, error)
	FindByIDRole(ctx context.Context, id uint64) (*Role, error)
	FindByName(ctx context.Context, name string) (*Role, error)
	CreateRole(ctx context.Context, r *Role) (uint64, error)
	UpdateRole(ctx context.Context, r *Role) error
	DeleteRole(ctx context.Context, id uint64) error
	GetRolesByUserID(ctx context.Context, userID uint64) ([]Role, error)
	GetRoleIDsByUserID(ctx context.Context, userID uint64) ([]uint64, error)
	AssignRolePermissions(ctx context.Context, roleID uint64, permKeys []string) error
	GetRolePermKeys(ctx context.Context, roleID uint64) ([]string, error)
	AssignUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error
	IsAdminRole(roleName string) bool
}
