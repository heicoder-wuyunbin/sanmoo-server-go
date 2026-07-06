package mpuser

import (
	"context"

	dommpuser "sanmoo-server-go/internal/domain/mpuser"
	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/shared/pagination"
)

// Service 微信用户管理应用服务。
type Service struct {
	repo dommpuser.Repository
}

// NewService 创建服务实例。
func NewService(repo dommpuser.Repository) *Service {
	return &Service{repo: repo}
}

// ListMPUsers 分页查询微信用户列表。
func (s *Service) ListMPUsers(ctx context.Context, page, size int, keyword, tagName string) (*dto.PageResponse[dommpuser.MPUserSummary], error) {
	p := pagination.Normalize(page, size)
	list, total, err := s.repo.ListMPUsers(ctx, dommpuser.ListQuery{
		Page:    p.Page,
		Size:    p.Size,
		Keyword: keyword,
		TagName: tagName,
	})
	if err != nil {
		return nil, err
	}
	return &dto.PageResponse[dommpuser.MPUserSummary]{
		List:  list,
		Total: total,
		Page:  p.Page,
		Size:  p.Size,
	}, nil
}

// GetMPUserDetail 获取微信用户详情。
func (s *Service) GetMPUserDetail(ctx context.Context, openID string) (*dommpuser.MPUserDetail, error) {
	return s.repo.GetMPUserDetail(ctx, openID)
}

// GetMPUserProfile 获取用户六边形画像。
func (s *Service) GetMPUserProfile(ctx context.Context, openID string) (*dommpuser.UserProfile, error) {
	return s.repo.GetMPUserProfile(ctx, openID)
}

// GenerateUserProfile 基于行为数据生成用户六边形画像。
func (s *Service) GenerateUserProfile(ctx context.Context, openID string) (*dommpuser.UserProfile, error) {
	return s.repo.ComputeAndSaveProfile(ctx, openID)
}

// GenerateUserTags 基于行为数据自动打标签。
func (s *Service) GenerateUserTags(ctx context.Context, openID string) ([]dommpuser.UserTag, error) {
	return s.repo.ComputeAndSaveTags(ctx, openID)
}

// RefreshRadar 刷新雷达图（行为标签 + 兴趣维度 + 六边形画像）。
func (s *Service) RefreshRadar(ctx context.Context, openID string) (*dommpuser.RadarData, error) {
	return s.repo.ComputeAndSaveRadar(ctx, openID)
}

// DeleteUserTag 删除用户标签。
func (s *Service) DeleteUserTag(ctx context.Context, tagID uint64) error {
	return s.repo.DeleteMPUserTag(ctx, tagID)
}

// GetUserTags 获取用户标签列表。
func (s *Service) GetUserTags(ctx context.Context, openID string) ([]dommpuser.UserTag, error) {
	return s.repo.GetMPUserTags(ctx, openID)
}
