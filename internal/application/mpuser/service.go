package mpuser

import (
	"context"

	domarticle "sanmoo-server-go/internal/domain/article"
	dommpuser "sanmoo-server-go/internal/domain/mpuser"
	domuser "sanmoo-server-go/internal/domain/user"
	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/shared/pagination"
)

// Service 微信用户管理应用服务。
type Service struct {
	repo     dommpuser.Repository
	userRepo domuser.Repository
}

// NewService 创建服务实例。
func NewService(repo dommpuser.Repository, userRepo domuser.Repository) *Service {
	return &Service{repo: repo, userRepo: userRepo}
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
//
// FROZEN (L2): 该方法属于"重运营画像"能力，已于 L2 冻结。
// 小程序用户域已降级为"轻阅读用户能力"，不再主动计算六边形画像。
// 为避免前端直接崩溃，接口签名保留并返回空画像结构；
// 实际计算逻辑（ComputeAndSaveProfile）已短路，不再写入数据库。
// 详见 documents/mp-user-domain-downgrade.md。
func (s *Service) GenerateUserProfile(ctx context.Context, openID string) (*dommpuser.UserProfile, error) {
	return &dommpuser.UserProfile{
		Dimensions: []dommpuser.ProfileDimension{},
	}, nil
}

// GenerateUserTags 基于行为数据自动打标签。
//
// FROZEN (L2): 该方法属于"重运营标签"能力，已于 L2 冻结。
// 小程序用户域已降级为"轻阅读用户能力"，不再自动生成行为/兴趣标签。
// 为避免前端直接崩溃，接口签名保留并返回空标签列表；
// 实际计算逻辑（ComputeAndSaveTags）已短路，不再写入数据库。
// 详见 documents/mp-user-domain-downgrade.md。
func (s *Service) GenerateUserTags(ctx context.Context, openID string) ([]dommpuser.UserTag, error) {
	return []dommpuser.UserTag{}, nil
}

// RefreshRadar 刷新雷达图（行为标签 + 兴趣维度 + 六边形画像）。
//
// FROZEN (L2): 该方法属于"雷达画像生成"能力，已于 L2 冻结。
// 小程序用户域已降级为"轻阅读用户能力"，不再聚合生成雷达画像。
// 为避免前端直接崩溃，接口签名保留并返回空雷达结构；
// 实际计算逻辑（ComputeAndSaveRadar）已短路，不再写入数据库。
// 详见 documents/mp-user-domain-downgrade.md。
func (s *Service) RefreshRadar(ctx context.Context, openID string) (*dommpuser.RadarData, error) {
	return &dommpuser.RadarData{
		Tags:      []dommpuser.UserTag{},
		Interests: []dommpuser.UserInterest{},
		Profile: &dommpuser.UserProfile{
			Dimensions: []dommpuser.ProfileDimension{},
		},
	}, nil
}

// DeleteUserTag 删除用户标签。
func (s *Service) DeleteUserTag(ctx context.Context, tagID uint64) error {
	return s.repo.DeleteMPUserTag(ctx, tagID)
}

// GetUserTags 获取用户标签列表。
func (s *Service) GetUserTags(ctx context.Context, openID string) ([]dommpuser.UserTag, error) {
	return s.repo.GetMPUserTags(ctx, openID)
}

// ---- MP public methods (moved from user service) ----

// MPGetUserProfile 获取小程序用户基本信息。
func (s *Service) MPGetUserProfile(ctx context.Context, openID string) (*dto.MPUserProfileResponse, error) {
	nickname, avatar, err := s.userRepo.GetMPUserNicknameAvatar(ctx, openID)
	if err != nil {
		return nil, err
	}
	return &dto.MPUserProfileResponse{Nickname: nickname, Avatar: avatar}, nil
}

// MPUpdateUserProfile 更新小程序用户昵称和头像。
func (s *Service) MPUpdateUserProfile(ctx context.Context, openID, nickName, avatarUrl string) error {
	if err := s.userRepo.UpsertMPUser(ctx, openID); err != nil {
		return err
	}
	return s.userRepo.UpdateMPUserProfile(ctx, openID, nickName, avatarUrl)
}

// ReportMPBehavior 上报小程序用户行为。
func (s *Service) ReportMPBehavior(ctx context.Context, openID string, articleID uint64, eventType string, staySeconds int, scene, strategy string) error {
	typ := eventType
	if typ == "" {
		typ = "view"
	}
	return s.userRepo.RecordMPBehavior(ctx, openID, articleID, typ, staySeconds, scene, strategy)
}

// MPAddFavorite 添加收藏。
func (s *Service) MPAddFavorite(ctx context.Context, openID string, articleID uint64) error {
	return s.userRepo.AddMPFavorite(ctx, openID, articleID)
}

// MPRemoveFavorite 取消收藏。
func (s *Service) MPRemoveFavorite(ctx context.Context, openID string, articleID uint64) error {
	return s.userRepo.RemoveMPFavorite(ctx, openID, articleID)
}

// MPFavoriteStatus 查询收藏状态。
func (s *Service) MPFavoriteStatus(ctx context.Context, openID string, articleID uint64) (*dto.MPFavoriteStatusResponse, error) {
	favored, err := s.userRepo.IsMPFavorited(ctx, openID, articleID)
	if err != nil {
		return nil, err
	}
	return &dto.MPFavoriteStatusResponse{IsFavorited: favored}, nil
}

// MPFavoriteList 收藏列表。
func (s *Service) MPFavoriteList(ctx context.Context, openID string, page, size int) (*dto.PageResponse[domarticle.Article], error) {
	p := pagination.Normalize(page, size)
	list, total, err := s.userRepo.ListMPFavorites(ctx, openID, p.Page, p.Size)
	if err != nil {
		return nil, err
	}
	return &dto.PageResponse[domarticle.Article]{List: list, Total: total, Page: p.Page, Size: p.Size}, nil
}

// AddMPBrowseHistory 添加浏览历史。
func (s *Service) AddMPBrowseHistory(ctx context.Context, openID string, articleID uint64) error {
	return s.userRepo.AddMPBrowseHistory(ctx, openID, articleID)
}

// ClearMPBrowseHistory 清空浏览历史。
func (s *Service) ClearMPBrowseHistory(ctx context.Context, openID string) error {
	return s.userRepo.ClearMPBrowseHistory(ctx, openID)
}

// MPBrowseHistoryList 浏览历史列表。
func (s *Service) MPBrowseHistoryList(ctx context.Context, openID string, page, size int) (*dto.PageResponse[domarticle.Article], error) {
	p := pagination.Normalize(page, size)
	list, total, err := s.userRepo.ListMPBrowseHistory(ctx, openID, p.Page, p.Size)
	if err != nil {
		return nil, err
	}
	return &dto.PageResponse[domarticle.Article]{List: list, Total: total, Page: p.Page, Size: p.Size}, nil
}

// MPDeleteUser 删除小程序用户。
func (s *Service) MPDeleteUser(ctx context.Context, openID string) error {
	return s.userRepo.DeleteMPUser(ctx, openID)
}

// MPSetSubscribe 设置订阅状态。
func (s *Service) MPSetSubscribe(ctx context.Context, openID string, subscribe bool) error {
	return s.userRepo.SetMPUserSubscribe(ctx, openID, subscribe)
}

// MPGetSubscribe 获取订阅状态。
func (s *Service) MPGetSubscribe(ctx context.Context, openID string) (bool, error) {
	return s.userRepo.GetMPUserSubscribe(ctx, openID)
}
