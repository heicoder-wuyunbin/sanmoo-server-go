package mysqlrepo

import (
	"context"
	"time"

	domcategory "sanmoo-server-go/internal/domain/category"
)

func (r *Repository) ListAllWithCountCategory(ctx context.Context) ([]domcategory.Category, error) {
	// 使用 LEFT JOIN 一次性获取所有分类及其文章数量
	type categoryWithCount struct {
		ID           uint64    `gorm:"column:id"`
		Name         string    `gorm:"column:name"`
		ArticleCount int64     `gorm:"column:article_count"`
		CreateTime   time.Time `gorm:"column:create_time"`
		UpdateTime   time.Time `gorm:"column:update_time"`
	}

	var results []categoryWithCount
	err := r.db.WithContext(ctx).Table("t_category").
		Select("t_category.id, t_category.name, COUNT(t_article_category_rel.category_id) as article_count, t_category.create_time, t_category.update_time").
		Joins("LEFT JOIN t_article_category_rel ON t_category.id = t_article_category_rel.category_id").
		Where("t_category.deleted_at IS NULL").
		Group("t_category.id").
		Order("t_category.id desc").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	categories := make([]domcategory.Category, 0, len(results))
	for _, result := range results {
		categories = append(categories, domcategory.Category{
			ID:           result.ID,
			Name:         result.Name,
			ArticleCount: int(result.ArticleCount),
			CreateTime:   result.CreateTime,
			UpdateTime:   result.UpdateTime,
		})
	}
	return categories, nil
}
func (r *Repository) FindByIDCategory(ctx context.Context, id uint64) (*domcategory.Category, error) {
	var c tCategory
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&c, id).Error; err != nil {
		return nil, err
	}
	var n int64
	r.db.WithContext(ctx).Table("t_article_category_rel").Where("category_id = ?", c.ID).Count(&n)
	return &domcategory.Category{ID: c.ID, Name: c.Name, ArticleCount: int(n), CreateTime: c.CreateTime, UpdateTime: c.UpdateTime}, nil
}
func (r *Repository) ListCategoriesByIDs(ctx context.Context, ids []uint64) ([]domcategory.Category, error) {
	var cats []tCategory
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&cats).Error; err != nil {
		return nil, err
	}
	res := make([]domcategory.Category, 0, len(cats))
	for _, c := range cats {
		var n int64
		r.db.WithContext(ctx).Table("t_article_category_rel").Where("category_id = ?", c.ID).Count(&n)
		res = append(res, domcategory.Category{ID: c.ID, Name: c.Name, ArticleCount: int(n), CreateTime: c.CreateTime, UpdateTime: c.UpdateTime})
	}
	return res, nil
}
func (r *Repository) CreateCategory(ctx context.Context, c *domcategory.Category) (uint64, error) {
	m := tCategory{Name: c.Name}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}
func (r *Repository) UpdateCategory(ctx context.Context, c *domcategory.Category) error {
	return r.db.WithContext(ctx).Model(&tCategory{}).Where("id=? AND deleted_at IS NULL", c.ID).Update("name", c.Name).Error
}
func (r *Repository) DeleteCategory(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&tCategory{}).Where("id=? AND deleted_at IS NULL", id).Update("deleted_at", &now).Error
}
func (r *Repository) BatchDeleteCategory(ctx context.Context, ids []uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&tCategory{}).Where("id IN ? AND deleted_at IS NULL", ids).Update("deleted_at", &now).Error
}
