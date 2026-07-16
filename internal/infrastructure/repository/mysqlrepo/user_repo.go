package mysqlrepo

import (
	"context"

	domuser "sanmoo-server-go/internal/domain/user"
	apperr "sanmoo-server-go/internal/shared/errors"

	"gorm.io/gorm"
)

func (r *Repository) FindByIDUser(ctx context.Context, id uint64) (*domuser.User, error) {
	var model tUser
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	return tUserToUser(&model), nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*domuser.User, error) {
	var model tUser
	err := r.db.WithContext(ctx).Where("username = ?", username).Take(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	return tUserToUser(&model), nil
}

func tUserToUser(m *tUser) *domuser.User {
	return &domuser.User{
		ID:                m.ID,
		Username:          m.Username,
		PasswordHash:      m.PasswordHash,
		Email:             m.Email,
		Nickname:          m.Nickname,
		Avatar:            m.Avatar,
		Status:            m.Status,
		LastLoginTime:     m.LastLoginTime,
		LastLoginIp:       m.LastLoginIp,
		LoginFailureCount: m.LoginFailureCount,
		LockedUntil:       m.LockedUntil,
		IsAdmin:           m.IsAdmin,
		CreateTime:        m.CreateTime,
		UpdateTime:        m.UpdateTime,
	}
}

func (r *Repository) DeleteUser(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&tUser{}, id).Error
}

func (r *Repository) BatchDeleteUsers(ctx context.Context, ids []uint64) error {
	return r.db.WithContext(ctx).Where("id in ?", ids).Delete(&tUser{}).Error
}

func (r *Repository) ToggleUserStatus(ctx context.Context, id uint64) error {
	var user tUser
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return err
	}

	newStatus := "DISABLED"
	if user.Status == "DISABLED" {
		newStatus = "ENABLED"
	}

	return r.db.WithContext(ctx).Model(&tUser{}).Where("id = ?", id).Update("status", newStatus).Error
}

func (r *Repository) UpdateUser(ctx context.Context, u *domuser.User) error {
	updates := map[string]any{}
	if u.Username != "" {
		updates["username"] = u.Username
	}
	if u.Email != "" {
		updates["email"] = u.Email
	}
	if u.Nickname != "" {
		updates["nickname"] = u.Nickname
	}
	if u.Avatar != "" {
		updates["avatar"] = u.Avatar
	}
	if u.PasswordHash != "" {
		updates["password_hash"] = u.PasswordHash
	}
	if u.LastLoginTime != nil {
		updates["last_login_time"] = u.LastLoginTime
	}
	if u.LastLoginIp != "" {
		updates["last_login_ip"] = u.LastLoginIp
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&tUser{}).Where("id = ?", u.ID).Updates(updates).Error
}