package mysqlrepo

import (
	"context"
	"errors"
	"time"

	apperr "sanmoo-server-go/internal/shared/errors"
	domtag "sanmoo-server-go/internal/domain/tag"

	"gorm.io/gorm"
)

func (r *Repository) ListTags(ctx context.Context, q domtag.ListQuery) ([]domtag.Tag, int64, error) {
	db := r.db.WithContext(ctx).Table("t_tag").Where("deleted_at IS NULL")
	if q.Keyword != "" {
		db = db.Where("name like ?", "%"+q.Keyword+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tags []tTag
	query := r.db.WithContext(ctx).Table("t_tag").Where("deleted_at IS NULL")
	if q.Keyword != "" {
		query = query.Where("name like ?", "%"+q.Keyword+"%")
	}
	query = query.Order("id desc")
	if q.Page > 0 && q.Size > 0 {
		query = query.Offset((q.Page - 1) * q.Size).Limit(q.Size)
	}
	if err := query.Find(&tags).Error; err != nil {
		return nil, 0, err
	}
	res := make([]domtag.Tag, 0, len(tags))
	for _, t := range tags {
		var c int64
		r.db.WithContext(ctx).Table("t_article_tag_rel").Where("tag_id = ?", t.ID).Count(&c)
		res = append(res, domtag.Tag{ID: t.ID, Name: t.Name, ArticleCount: int(c), CreateTime: t.CreateTime, UpdateTime: t.UpdateTime})
	}
	return res, total, nil
}
func (r *Repository) ListAllWithCountTag(ctx context.Context) ([]domtag.Tag, error) {
	// 使用 LEFT JOIN 一次性获取所有标签及其文章数量
	type tagWithCount struct {
		ID           uint64    `gorm:"column:id"`
		Name         string    `gorm:"column:name"`
		ArticleCount int64     `gorm:"column:article_count"`
		CreateTime   time.Time `gorm:"column:create_time"`
		UpdateTime   time.Time `gorm:"column:update_time"`
	}

	var results []tagWithCount
	err := r.db.WithContext(ctx).Table("t_tag").
		Select("t_tag.id, t_tag.name, COUNT(t_article_tag_rel.tag_id) as article_count, t_tag.create_time, t_tag.update_time").
		Joins("LEFT JOIN t_article_tag_rel ON t_tag.id = t_article_tag_rel.tag_id").
		Where("t_tag.deleted_at IS NULL").
		Group("t_tag.id").
		Order("t_tag.id desc").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	tags := make([]domtag.Tag, 0, len(results))
	for _, result := range results {
		tags = append(tags, domtag.Tag{
			ID:           result.ID,
			Name:         result.Name,
			ArticleCount: int(result.ArticleCount),
			CreateTime:   result.CreateTime,
			UpdateTime:   result.UpdateTime,
		})
	}
	return tags, nil
}
func (r *Repository) FindByIDTag(ctx context.Context, id uint64) (*domtag.Tag, error) {
	var t tTag
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&t, id).Error; err != nil {
		return nil, err
	}
	var c int64
	r.db.WithContext(ctx).Table("t_article_tag_rel").Where("tag_id = ?", id).Count(&c)
	return &domtag.Tag{ID: t.ID, Name: t.Name, ArticleCount: int(c), CreateTime: t.CreateTime, UpdateTime: t.UpdateTime}, nil
}
func (r *Repository) ListTagsByIDs(ctx context.Context, ids []uint64) ([]domtag.Tag, error) {
	var tags []tTag
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&tags).Error; err != nil {
		return nil, err
	}
	res := make([]domtag.Tag, 0, len(tags))
	for _, t := range tags {
		var c int64
		r.db.WithContext(ctx).Table("t_article_tag_rel").Where("tag_id = ?", t.ID).Count(&c)
		res = append(res, domtag.Tag{ID: t.ID, Name: t.Name, ArticleCount: int(c), CreateTime: t.CreateTime, UpdateTime: t.UpdateTime})
	}
	return res, nil
}
func (r *Repository) CreateTag(ctx context.Context, t *domtag.Tag) (uint64, error) {
	m := tTag{Name: t.Name}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}
func (r *Repository) UpdateTag(ctx context.Context, t *domtag.Tag) error {
	// 检查同名标签是否已存在（排除当前标签 + 未软删除）
	var dup tTag
	err := r.db.WithContext(ctx).Model(&tTag{}).Where("name=? AND id!=? AND deleted_at IS NULL", t.Name, t.ID).Take(&dup).Error
	if err == nil {
		return apperr.New(apperr.ErrConflict.Code, "标签名称已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.WithContext(ctx).Model(&tTag{}).Where("id=? AND deleted_at IS NULL", t.ID).Update("name", t.Name).Error
}
func (r *Repository) DeleteTag(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&tTag{}).Where("id=? AND deleted_at IS NULL", id).Update("deleted_at", &now).Error
}
func (r *Repository) BatchDeleteTag(ctx context.Context, ids []uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&tTag{}).Where("id IN ? AND deleted_at IS NULL", ids).Update("deleted_at", &now).Error
}
