package mpuser

import "context"

// ListQuery 微信用户列表查询条件。
type ListQuery struct {
	Page    int
	Size    int
	Keyword string
	TagName string
}

// Repository 微信用户管理仓库接口。
type Repository interface {
	// ListMPUsers 分页查询微信用户列表。
	ListMPUsers(ctx context.Context, q ListQuery) ([]MPUserSummary, int64, error)

	// GetMPUserDetail 获取微信用户详情。
	GetMPUserDetail(ctx context.Context, openID string) (*MPUserDetail, error)

	// GetMPUserTags 获取用户标签列表。
	GetMPUserTags(ctx context.Context, openID string) ([]UserTag, error)

	// UpsertMPUserTag 新增或更新用户标签。
	UpsertMPUserTag(ctx context.Context, openID, tagName, tagCategory string, score float64, source string) error

	// DeleteMPUserTag 删除用户标签。
	DeleteMPUserTag(ctx context.Context, tagID uint64) error

	// GetMPUserProfile 获取用户六边形画像。
	GetMPUserProfile(ctx context.Context, openID string) (*UserProfile, error)

	// SaveMPUserProfile 保存用户画像维度得分。
	SaveMPUserProfile(ctx context.Context, openID, dimension string, score float64) error

	// ComputeAndSaveProfile 基于行为数据计算并保存用户六边形画像。
	ComputeAndSaveProfile(ctx context.Context, openID string) (*UserProfile, error)

	// ComputeAndSaveTags 基于行为数据自动打标签。
	ComputeAndSaveTags(ctx context.Context, openID string) ([]UserTag, error)

	// ComputeAndSaveRadar 刷新雷达图数据（行为标签 + 兴趣维度 + 六边形画像）。
	ComputeAndSaveRadar(ctx context.Context, openID string) (*RadarData, error)

	// GetMPUserInterests 获取用户兴趣维度（含名称）。
	GetMPUserInterests(ctx context.Context, openID string) ([]UserInterest, error)

	// CountMPUserBehavior 统计用户行为次数。
	CountMPUserBehavior(ctx context.Context, openID string) (viewCount int64, favoriteCount int64, totalStaySeconds int64, err error)
}
