package mysqlrepo

import (
	"context"
	"time"

	dommpuser "sanmoo-server-go/internal/domain/mpuser"

	"gorm.io/gorm"
)

// ======================== MP User Admin Repository ========================

// ListMPUsers 分页查询微信用户列表（含浏览数、收藏数统计，轻运营：移除标签数）。
func (r *Repository) ListMPUsers(ctx context.Context, q dommpuser.ListQuery) ([]dommpuser.MPUserSummary, int64, error) {
	var total int64
	db := r.db.WithContext(ctx).Table("t_mp_user")

	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		db = db.Where("nickname LIKE ? OR openid LIKE ?", kw, kw)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		ID             uint64    `gorm:"column:id"`
		OpenID         string    `gorm:"column:openid"`
		Nickname       string    `gorm:"column:nickname"`
		Avatar         string    `gorm:"column:avatar"`
		Status         uint8     `gorm:"column:status"`
		FirstLoginTime time.Time `gorm:"column:first_login_time"`
		LastLoginTime  time.Time `gorm:"column:last_login_time"`
		CreateTime     time.Time `gorm:"column:create_time"`
		ViewCount      int64     `gorm:"column:view_count"`
		FavoriteCount  int64     `gorm:"column:favorite_count"`
	}

	var rows []row
	err := db.Select(`id,
		openid,
		nickname,
		avatar,
		status,
		first_login_time,
		last_login_time,
		create_time,
		COALESCE((SELECT COUNT(1) FROM t_mp_browse_history bh WHERE bh.openid = t_mp_user.openid), 0) AS view_count,
		COALESCE((SELECT COUNT(1) FROM t_mp_user_favorite WHERE openid = t_mp_user.openid), 0) AS favorite_count`).
		Order("last_login_time DESC").
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]dommpuser.MPUserSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, dommpuser.MPUserSummary{
			ID:             row.ID,
			OpenID:         row.OpenID,
			Nickname:       row.Nickname,
			Avatar:         row.Avatar,
			Status:         row.Status,
			FirstLoginTime: row.FirstLoginTime,
			LastLoginTime:  row.LastLoginTime,
			ViewCount:      row.ViewCount,
			FavoriteCount:  row.FavoriteCount,
			CreateTime:     row.CreateTime,
		})
	}
	return out, total, nil
}