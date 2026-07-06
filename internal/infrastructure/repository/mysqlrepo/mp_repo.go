package mysqlrepo

import (
	"context"
	"errors"
	"strings"
	"time"

	domarticle "sanmoo-server-go/internal/domain/article"
	apperr "sanmoo-server-go/internal/shared/errors"

	"gorm.io/gorm"
)

func (r *Repository) UpsertMPUser(ctx context.Context, openID string) error {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return gorm.ErrInvalidData
	}
	now := time.Now()
	return r.db.WithContext(ctx).Exec(`
INSERT INTO t_mp_user(openid, status, first_login_time, last_login_time)
VALUES(?, 1, ?, ?)
ON DUPLICATE KEY UPDATE last_login_time = VALUES(last_login_time), status = 1
`, openID, now, now).Error
}

func (r *Repository) GetMPUserNicknameAvatar(ctx context.Context, openID string) (string, string, error) {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return "", "", apperr.ErrInvalidParam
	}
	var out struct {
		Nickname string `gorm:"column:nickname"`
		Avatar   string `gorm:"column:avatar"`
	}
	err := r.db.WithContext(ctx).Table("t_mp_user").Select("nickname, avatar").Where("openid = ?", openID).Take(&out).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", nil
		}
		return "", "", err
	}
	return out.Nickname, out.Avatar, nil
}

func (r *Repository) UpdateMPUserProfile(ctx context.Context, openID, nickname, avatar string) error {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return gorm.ErrInvalidData
	}
	return r.db.WithContext(ctx).Exec(`
UPDATE t_mp_user
SET nickname = ?, avatar = ?, update_time = CURRENT_TIMESTAMP
WHERE openid = ?
`, strings.TrimSpace(nickname), strings.TrimSpace(avatar), openID).Error
}

func (r *Repository) RecordMPBehavior(ctx context.Context, openID string, articleID uint64, eventType string, staySeconds int, scene, strategy string) error {
	return r.db.WithContext(ctx).Create(&tMPUserBehavior{
		OpenID:      strings.TrimSpace(openID),
		ArticleID:   articleID,
		EventType:   strings.ToLower(strings.TrimSpace(eventType)),
		StaySeconds: staySeconds,
		Scene:       strings.TrimSpace(scene),
		Strategy:    strings.TrimSpace(strategy),
		EventTime:   time.Now(),
	}).Error
}

func (r *Repository) AddMPInterest(ctx context.Context, openID, dimensionType string, dimensionID uint64, delta float64) error {
	return r.db.WithContext(ctx).Exec(`
INSERT INTO t_mp_user_interest(openid, dimension_type, dimension_id, score)
VALUES(?, ?, ?, ?)
ON DUPLICATE KEY UPDATE score = score + VALUES(score), update_time = CURRENT_TIMESTAMP
`, strings.TrimSpace(openID), strings.TrimSpace(dimensionType), dimensionID, delta).Error
}

func (r *Repository) AddMPFavorite(ctx context.Context, openID string, articleID uint64) error {
	return r.db.WithContext(ctx).Exec(`
INSERT INTO t_mp_user_favorite(openid, article_id)
VALUES(?, ?)
ON DUPLICATE KEY UPDATE update_time = CURRENT_TIMESTAMP
`, strings.TrimSpace(openID), articleID).Error
}

func (r *Repository) RemoveMPFavorite(ctx context.Context, openID string, articleID uint64) error {
	return r.db.WithContext(ctx).Where("openid = ? AND article_id = ?", strings.TrimSpace(openID), articleID).Delete(&tMPFavorite{}).Error
}

func (r *Repository) IsMPFavorited(ctx context.Context, openID string, articleID uint64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("t_mp_user_favorite").Where("openid = ? AND article_id = ?", strings.TrimSpace(openID), articleID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ListMPFavorites(ctx context.Context, openID string, page, size int) ([]domarticle.Article, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Table("t_mp_user_favorite").
		Where("openid = ?", strings.TrimSpace(openID)).
		Distinct("article_id").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	subQuery := r.db.WithContext(ctx).Table("t_mp_user_favorite").
		Select("article_id, MAX(update_time) as update_time").
		Where("openid = ?", strings.TrimSpace(openID)).
		Group("article_id")
	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Table("(?) as f", subQuery).
			Joins("join t_article a on a.id = f.article_id").
			Joins("left join (select article_id, min(category_id) as category_id from t_article_category_rel group by article_id) acr on acr.article_id = a.id").
			Joins("left join t_category c on c.id = acr.category_id")
	}
	type row struct {
		ID                                           uint64
		Title, TitleImage, Description, CategoryName string
		CategoryID                                   uint64
		ReadNum                                      int
		IsTop, IsPublished                           bool
		CreateTime, UpdateTime                       time.Time
	}
	var rows []row
	err := buildBase().
		Select("a.id, a.title, a.title_image, a.description, a.read_num, a.is_top, a.is_published, a.create_time, a.update_time, acr.category_id, c.name as category_name").
		Order("f.update_time desc").
		Offset((page - 1) * size).Limit(size).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	res := make([]domarticle.Article, 0, len(rows))
	for _, row := range rows {
		tags, _ := r.queryArticleTags(ctx, row.ID)
		res = append(res, domarticle.Article{
			ID:          row.ID,
			Title:       row.Title,
			TitleImage:  row.TitleImage,
			Description: row.Description,
			ReadNum:     row.ReadNum,
			IsTop:       row.IsTop,
			IsPublished: row.IsPublished,
			CategoryID:  row.CategoryID,
			Category:    row.CategoryName,
			Tags:        tags,
			CreateTime:  row.CreateTime,
			UpdateTime:  row.UpdateTime,
		})
	}
	return res, total, nil
}

type tMPBrowseHistory struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	OpenID     string    `gorm:"column:openid"`
	ArticleID  uint64    `gorm:"column:article_id"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func (tMPBrowseHistory) TableName() string { return "t_mp_browse_history" }

func (r *Repository) AddMPBrowseHistory(ctx context.Context, openID string, articleID uint64) error {
	return r.db.WithContext(ctx).Exec(`
INSERT INTO t_mp_browse_history(openid, article_id)
VALUES(?, ?)
ON DUPLICATE KEY UPDATE update_time = CURRENT_TIMESTAMP
`, strings.TrimSpace(openID), articleID).Error
}

func (r *Repository) ClearMPBrowseHistory(ctx context.Context, openID string) error {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return gorm.ErrInvalidData
	}
	return r.db.WithContext(ctx).Where("openid = ?", openID).Delete(&tMPBrowseHistory{}).Error
}

func (r *Repository) SetMPUserSubscribe(ctx context.Context, openID string, subscribe bool) error {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return gorm.ErrInvalidData
	}
	return r.db.WithContext(ctx).Exec(`
INSERT INTO t_mp_user_subscribe(openid, subscribe)
VALUES(?, ?)
ON DUPLICATE KEY UPDATE subscribe = VALUES(subscribe), update_time = CURRENT_TIMESTAMP
`, openID, subscribe).Error
}

func (r *Repository) GetMPUserSubscribe(ctx context.Context, openID string) (bool, error) {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return false, apperr.ErrInvalidParam
	}
	var subscribe bool
	err := r.db.WithContext(ctx).Table("t_mp_user_subscribe").Select("subscribe").Where("openid = ?", openID).Take(&subscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return subscribe, nil
}

func (r *Repository) ListSubscribedOpenIDs(ctx context.Context) ([]string, error) {
	var openIDs []string
	err := r.db.WithContext(ctx).Table("t_mp_user_subscribe").Select("openid").Where("subscribe = ?", true).Scan(&openIDs).Error
	if err != nil {
		return nil, err
	}
	return openIDs, nil
}

func (r *Repository) DeleteMPUser(ctx context.Context, openID string) error {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return gorm.ErrInvalidData
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("openid = ?", openID).Delete(&tMPUser{}).Error; err != nil {
			return err
		}
		if err := tx.Where("openid = ?", openID).Delete(&tMPUserBehavior{}).Error; err != nil {
			return err
		}
		if err := tx.Where("openid = ?", openID).Delete(&tMPFavorite{}).Error; err != nil {
			return err
		}
		if err := tx.Where("openid = ?", openID).Delete(&tMPBrowseHistory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("openid = ?", openID).Delete(&tMPUserInterest{}).Error; err != nil {
			return err
		}
		if err := tx.Where("openid = ?", openID).Delete(&tMPUserTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("openid = ?", openID).Delete(&tMPUserProfile{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) ListMPBrowseHistory(ctx context.Context, openID string, page, size int) ([]domarticle.Article, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Table("t_mp_browse_history").
		Where("openid = ?", strings.TrimSpace(openID)).
		Distinct("article_id").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	subQuery := r.db.WithContext(ctx).Table("t_mp_browse_history").
		Select("article_id, MAX(update_time) as update_time").
		Where("openid = ?", strings.TrimSpace(openID)).
		Group("article_id")
	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Table("(?) as h", subQuery).
			Joins("join t_article a on a.id = h.article_id").
			Joins("left join (select article_id, min(category_id) as category_id from t_article_category_rel group by article_id) acr on acr.article_id = a.id").
			Joins("left join t_category c on c.id = acr.category_id")
	}
	type row struct {
		ID                                           uint64
		Title, TitleImage, Description, CategoryName string
		CategoryID                                   uint64
		ReadNum                                      int
		IsTop, IsPublished                           bool
		CreateTime, UpdateTime                       time.Time
	}
	var rows []row
	err := buildBase().
		Select("a.id, a.title, a.title_image, a.description, a.read_num, a.is_top, a.is_published, a.create_time, a.update_time, acr.category_id, c.name as category_name").
		Order("h.update_time desc").
		Offset((page - 1) * size).Limit(size).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	res := make([]domarticle.Article, 0, len(rows))
	for _, row := range rows {
		tags, _ := r.queryArticleTags(ctx, row.ID)
		res = append(res, domarticle.Article{
			ID:          row.ID,
			Title:       row.Title,
			TitleImage:  row.TitleImage,
			Description: row.Description,
			ReadNum:     row.ReadNum,
			IsTop:       row.IsTop,
			IsPublished: row.IsPublished,
			CategoryID:  row.CategoryID,
			Category:    row.CategoryName,
			Tags:        tags,
			CreateTime:  row.CreateTime,
			UpdateTime:  row.UpdateTime,
		})
	}
	return res, total, nil
}
