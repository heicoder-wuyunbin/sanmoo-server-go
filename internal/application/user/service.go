package user

import (
	"context"
	domarticle "sanmoo-server-go/internal/domain/article"
	domuser "sanmoo-server-go/internal/domain/user"
	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/shared/pagination"
)

type Service struct {
	repo domuser.Repository
}

func NewService(repo domuser.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUsers(ctx context.Context, page, size int, keyword string) (*pagination.PageData, error) {
	list, total, err := s.repo.ListUsers(ctx, domuser.ListQuery{Page: page, Size: size, Keyword: keyword})
	if err != nil {
		return nil, err
	}
	return pagination.NewPageData(list, total, page, size), nil
}

func (s *Service) CreateUser(ctx context.Context, username, password string, roleID uint64, email string) (uint64, error) {
	u := domuser.User{Username: username, RoleID: roleID}
	if email != "" {
		u.Email = email
	}
	if password != "" {
		if err := u.SetPassword(password); err != nil {
			return 0, err
		}
	}
	return s.repo.CreateUser(ctx, &u)
}

func (s *Service) UpdateUser(ctx context.Context, id uint64, username, password string, roleID uint64, email string) error {
	u, err := s.repo.FindByIDUser(ctx, id)
	if err != nil {
		return err
	}
	if email != "" {
		u.Email = email
	}
	u.Username = username
	u.RoleID = roleID
	if password != "" {
		if err := u.SetPassword(password); err != nil {
			return err
		}
	}
	return s.repo.UpdateUser(ctx, u)
}

func (s *Service) DeleteUser(ctx context.Context, id uint64) error { return s.repo.DeleteUser(ctx, id) }

func (s *Service) BatchDeleteUsers(ctx context.Context, ids []uint64) error {
	return s.repo.BatchDeleteUsers(ctx, ids)
}
func (s *Service) UpdateUserPassword(ctx context.Context, id uint64, password string) error {
	u, err := s.repo.FindByIDUser(ctx, id)
	if err != nil {
		return err
	}
	if err := u.SetPassword(password); err != nil {
		return err
	}
	return s.repo.UpdateUser(ctx, u)
}
func (s *Service) ToggleUserStatus(ctx context.Context, id uint64) error {
	return s.repo.ToggleUserStatus(ctx, id)
}
func (s *Service) ChangePassword(ctx context.Context, id uint64, oldPassword, newPassword string) error {
	return s.UpdateUserPassword(ctx, id, newPassword)
}

// MP methods
func (s *Service) MPGetUserProfile(ctx context.Context, openID string) (*dto.MPUserProfileResponse, error) {
	nickname, avatar, err := s.repo.GetMPUserNicknameAvatar(ctx, openID)
	if err != nil {
		return nil, err
	}
	return &dto.MPUserProfileResponse{Nickname: nickname, Avatar: avatar}, nil
}
func (s *Service) MPUpdateUserProfile(ctx context.Context, openID, nickName, avatarUrl string) error {
	if err := s.repo.UpsertMPUser(ctx, openID); err != nil {
		return err
	}
	return s.repo.UpdateMPUserProfile(ctx, openID, nickName, avatarUrl)
}
func (s *Service) ReportMPBehavior(ctx context.Context, openID string, articleID uint64, eventType string, staySeconds int, scene, strategy string) error {
	typ := eventType
	if typ == "" {
		typ = "view"
	}
	return s.repo.RecordMPBehavior(ctx, openID, articleID, typ, staySeconds, scene, strategy)
}
func (s *Service) MPAddFavorite(ctx context.Context, openID string, articleID uint64) error {
	return s.repo.AddMPFavorite(ctx, openID, articleID)
}
func (s *Service) MPRemoveFavorite(ctx context.Context, openID string, articleID uint64) error {
	return s.repo.RemoveMPFavorite(ctx, openID, articleID)
}
func (s *Service) MPFavoriteStatus(ctx context.Context, openID string, articleID uint64) (*dto.MPFavoriteStatusResponse, error) {
	favored, err := s.repo.IsMPFavorited(ctx, openID, articleID)
	if err != nil {
		return nil, err
	}
	return &dto.MPFavoriteStatusResponse{IsFavorited: favored}, nil
}
func (s *Service) MPFavoriteList(ctx context.Context, openID string, page, size int) (*dto.PageResponse[domarticle.Article], error) {
	p := pagination.Normalize(page, size)
	list, total, err := s.repo.ListMPFavorites(ctx, openID, p.Page, p.Size)
	if err != nil {
		return nil, err
	}
	return &dto.PageResponse[domarticle.Article]{List: list, Total: total, Page: p.Page, Size: p.Size}, nil
}
func (s *Service) AddMPBrowseHistory(ctx context.Context, openID string, articleID uint64) error {
	return s.repo.AddMPBrowseHistory(ctx, openID, articleID)
}
func (s *Service) MPBrowseHistoryList(ctx context.Context, openID string, page, size int) (*dto.PageResponse[domarticle.Article], error) {
	p := pagination.Normalize(page, size)
	list, total, err := s.repo.ListMPBrowseHistory(ctx, openID, p.Page, p.Size)
	if err != nil {
		return nil, err
	}
	return &dto.PageResponse[domarticle.Article]{List: list, Total: total, Page: p.Page, Size: p.Size}, nil
}
func (s *Service) ClearMPBrowseHistory(ctx context.Context, openID string) error {
	return s.repo.ClearMPBrowseHistory(ctx, openID)
}
func (s *Service) MPDeleteUser(ctx context.Context, openID string) error {
	return s.repo.DeleteMPUser(ctx, openID)
}
func (s *Service) MPSetSubscribe(ctx context.Context, openID string, subscribe bool) error {
	return s.repo.SetMPUserSubscribe(ctx, openID, subscribe)
}
func (s *Service) MPGetSubscribe(ctx context.Context, openID string) (bool, error) {
	return s.repo.GetMPUserSubscribe(ctx, openID)
}
func (s *Service) MPListSubscribedOpenIDs(ctx context.Context) ([]string, error) {
	return s.repo.ListSubscribedOpenIDs(ctx)
}
