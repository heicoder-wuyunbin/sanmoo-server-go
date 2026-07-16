package user

import (
	"context"
	domarticle "sanmoo-server-go/internal/domain/article"
)

type Repository interface {
	FindByIDUser(ctx context.Context, id uint64) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, u *User) error
	DeleteUser(ctx context.Context, id uint64) error
	BatchDeleteUsers(ctx context.Context, ids []uint64) error
	ToggleUserStatus(ctx context.Context, id uint64) error

	// --- Mini Program (MP) related ---
	UpsertMPUser(ctx context.Context, openID string) error
	GetMPUserNicknameAvatar(ctx context.Context, openID string) (nickname string, avatar string, err error)
	UpdateMPUserProfile(ctx context.Context, openID, nickname, avatar string) error
	RecordMPBehavior(ctx context.Context, openID string, articleID uint64, eventType string, staySeconds int, scene, strategy string) error

	AddMPFavorite(ctx context.Context, openID string, articleID uint64) error
	RemoveMPFavorite(ctx context.Context, openID string, articleID uint64) error
	IsMPFavorited(ctx context.Context, openID string, articleID uint64) (bool, error)
	ListMPFavorites(ctx context.Context, openID string, page, size int) ([]domarticle.Article, int64, error)

	AddMPBrowseHistory(ctx context.Context, openID string, articleID uint64) error
	ListMPBrowseHistory(ctx context.Context, openID string, page, size int) ([]domarticle.Article, int64, error)
	ClearMPBrowseHistory(ctx context.Context, openID string) error
	DeleteMPUser(ctx context.Context, openID string) error
	SetMPUserSubscribe(ctx context.Context, openID string, subscribe bool) error
	GetMPUserSubscribe(ctx context.Context, openID string) (bool, error)
	ListSubscribedOpenIDs(ctx context.Context) ([]string, error)
}
