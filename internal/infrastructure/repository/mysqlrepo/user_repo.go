package mysqlrepo

import (
	"context"
	"time"

	domuser "sanmoo-server-go/internal/domain/user"
	apperr "sanmoo-server-go/internal/shared/errors"

	"gorm.io/gorm"
)

func (r *Repository) FindByIDUser(ctx context.Context, id uint64) (*domuser.User, error) {
	tuple, err := r.queryUserTuple(r.db.WithContext(ctx).Where("u.id = ?", id))
	if err != nil {
		return nil, err
	}
	if tuple.ID == 0 {
		return nil, apperr.ErrNotFound
	}
	return tupleToUser(tuple), nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*domuser.User, error) {
	tuple, err := r.queryUserTuple(r.db.WithContext(ctx).Where("u.username = ?", username))
	if err != nil {
		return nil, err
	}
	if tuple.ID == 0 {
		return nil, apperr.ErrNotFound
	}
	return tupleToUser(tuple), nil
}

type userTuple struct {
	ID            uint64
	Username      string
	PasswordHash  string
	Email         string
	Nickname      string
	Avatar        string
	Status        string
	LastLoginTime *time.Time
	LastLoginIp   string
	CreateTime    time.Time
	UpdateTime    time.Time
	RoleID        uint64
	RoleName      string
}

func (r *Repository) queryUserTuple(base *gorm.DB) (*userTuple, error) {
	var out userTuple
	err := base.Table("t_user u").
		Select("u.id, u.username, u.password_hash, u.email, u.nickname, u.avatar, u.status, u.last_login_time, u.last_login_ip, u.create_time, u.update_time, ur.role_id, r.name as role_name").
		Joins("left join t_user_role ur on ur.user_id = u.id").
		Joins("left join t_role r on r.id = ur.role_id").
		Limit(1).Scan(&out).Error
	return &out, err
}
func tupleToUser(t *userTuple) *domuser.User {
	return &domuser.User{
		ID:            t.ID,
		Username:      t.Username,
		PasswordHash:  t.PasswordHash,
		Email:         t.Email,
		Nickname:      t.Nickname,
		Avatar:        t.Avatar,
		Status:        t.Status,
		LastLoginTime: t.LastLoginTime,
		LastLoginIp:   t.LastLoginIp,
		RoleID:        t.RoleID,
		RoleName:      t.RoleName,
		CreateTime:    t.CreateTime,
		UpdateTime:    t.UpdateTime,
	}
}

func (r *Repository) ListUsers(ctx context.Context, q domuser.ListQuery) ([]domuser.User, int64, error) {
	query := r.db.WithContext(ctx).Table("t_user u").Joins("left join t_user_role ur on ur.user_id = u.id").Joins("left join t_role r on r.id = ur.role_id")
	if q.Keyword != "" {
		query = query.Where("u.username like ?", "%"+q.Keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []userTuple
	err := query.Select("u.id, u.username, u.password_hash, u.email, u.nickname, u.avatar, u.status, u.last_login_time, u.last_login_ip, u.create_time, u.update_time, ur.role_id, r.name as role_name").
		Order("u.id desc").Offset((q.Page - 1) * q.Size).Limit(q.Size).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	res := make([]domuser.User, 0, len(rows))
	for _, row := range rows {
		u := tupleToUser(&row)
		res = append(res, *u)
	}
	return res, total, nil
}

func (r *Repository) CreateUser(ctx context.Context, u *domuser.User) (uint64, error) {
	model := tUser{Username: u.Username, PasswordHash: u.PasswordHash, Email: u.Email, Nickname: u.Nickname, Avatar: u.Avatar}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return 0, err
	}
	if err := r.db.WithContext(ctx).Table("t_user_role").Create(map[string]any{"user_id": model.ID, "role_id": u.RoleID, "created_by": "system", "updated_by": "system"}).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *Repository) UpdateUser(ctx context.Context, u *domuser.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{"username": u.Username}
		if u.Email != "" {
			updates["email"] = u.Email
		}
		if u.Nickname != "" {
			updates["nickname"] = u.Nickname
		}
		if u.Avatar != "" {
			updates["avatar"] = u.Avatar
		}
		if err := tx.Model(&tUser{}).Where("id = ?", u.ID).Updates(updates).Error; err != nil {
			return err
		}
		if u.PasswordHash != "" {
			if err := tx.Model(&tUser{}).Where("id = ?", u.ID).Update("password_hash", u.PasswordHash).Error; err != nil {
				return err
			}
		}
		// 只有当角色 ID 大于 0 时才更新
		if u.RoleID > 0 {
			if err := tx.Table("t_user_role").Where("user_id = ?", u.ID).Update("role_id", u.RoleID).Error; err != nil {
				return err
			}
		}
		// 登录信息（非必填）
		if u.LastLoginTime != nil {
			if err := tx.Model(&tUser{}).Where("id = ?", u.ID).Update("last_login_time", u.LastLoginTime).Error; err != nil {
				return err
			}
		}
		if u.LastLoginIp != "" {
			if err := tx.Model(&tUser{}).Where("id = ?", u.ID).Update("last_login_ip", u.LastLoginIp).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
