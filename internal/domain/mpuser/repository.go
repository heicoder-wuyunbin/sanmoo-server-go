package mpuser

import "context"

// ListQuery 微信用户列表查询条件（轻运营：移除 TagName 过滤）。
type ListQuery struct {
	Page    int
	Size    int
	Keyword string
}

// Repository 微信用户管理仓库接口（轻运营：仅保留列表查询）。
type Repository interface {
	// ListMPUsers 分页查询微信用户列表。
	ListMPUsers(ctx context.Context, q ListQuery) ([]MPUserSummary, int64, error)
}