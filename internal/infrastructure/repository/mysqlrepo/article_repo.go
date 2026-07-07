package mysqlrepo

import (
	"context"
	"sort"
	"time"

	domarticle "sanmoo-server-go/internal/domain/article"
	"gorm.io/gorm"
)

type articleListRow struct {
	ID                                           uint64
	Title, Slug, TitleImage, Description, CategoryName string
	CategoryID                                   uint64
	ReadNum, LikeNum                             int
	IsTop, IsPublished                           bool
	CreateTime, UpdateTime                       time.Time
}

const articleListSelectSQL = `a.id, a.title, a.slug, a.title_image, a.description, a.read_num, a.like_num,
		a.is_top, a.is_published, a.create_time, a.update_time,
		acr.category_id, c.name as category_name`

const articleCategoryJoinSQL = `left join (
		select article_id, min(category_id) as category_id
		from t_article_category_rel
		group by article_id
	) acr on acr.article_id = a.id
	left join t_category c on c.id = acr.category_id`

func mapArticleRow(row articleListRow, tagsMap map[uint64][]domarticle.TagRef) domarticle.Article {
	return domarticle.Article{
		ID:          row.ID,
		Title:       row.Title,
		Slug:        row.Slug,
		TitleImage:  row.TitleImage,
		Description: row.Description,
		ReadNum:     row.ReadNum,
		LikeNum:     row.LikeNum,
		IsTop:       row.IsTop,
		IsPublished: row.IsPublished,
		CategoryID:  row.CategoryID,
		Category:    row.CategoryName,
		Tags:        tagsMap[row.ID],
		CreateTime:  row.CreateTime,
		UpdateTime:  row.UpdateTime,
	}
}

func (r *Repository) batchMapArticles(ctx context.Context, rows []articleListRow) []domarticle.Article {
	articleIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		articleIDs = append(articleIDs, row.ID)
	}
	tagsMap := r.batchQueryArticleTags(ctx, articleIDs)

	res := make([]domarticle.Article, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapArticleRow(row, tagsMap))
	}
	return res
}

func (r *Repository) ListArticles(ctx context.Context, q domarticle.ListQuery) ([]domarticle.Article, int64, error) {
	base := r.db.WithContext(ctx).Table("t_article a")
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		base = base.Where(
			`a.title like ? or a.description like ? or exists (
				select 1
				from t_article_tag_rel atr
				join t_tag t on t.id = atr.tag_id
				where atr.article_id = a.id and t.name like ?
			)`,
			kw, kw, kw,
		)
	}
	if q.CategoryID > 0 {
		base = base.Where("exists (select 1 from t_article_category_rel acrf where acrf.article_id = a.id and acrf.category_id = ?)", q.CategoryID)
	}
	if q.TagID > 0 {
		base = base.Where("exists (select 1 from t_article_tag_rel atrf where atrf.article_id = a.id and atrf.tag_id = ?)", q.TagID)
	}
	if q.IsPublished != nil {
		base = base.Where("a.is_published=?", *q.IsPublished)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []articleListRow
	err := base.
		Select(articleListSelectSQL).
		Joins(articleCategoryJoinSQL).
		Order("a.is_top desc, a.create_time desc").
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return r.batchMapArticles(ctx, rows), total, nil
}

// ListArticle 兼容推荐策略模块的查询接口（推荐模块使用单数命名）。
func (r *Repository) ListArticle(ctx context.Context, q domarticle.ListQuery) ([]domarticle.Article, int64, error) {
	return r.ListArticles(ctx, q)
}

func (r *Repository) ListArticlesByIDs(ctx context.Context, ids []uint64) ([]domarticle.Article, error) {
	if len(ids) == 0 {
		return []domarticle.Article{}, nil
	}

	var rows []articleListRow
	err := r.db.WithContext(ctx).
		Table("t_article a").
		Select(articleListSelectSQL).
		Joins(articleCategoryJoinSQL).
		Where("a.id in (?)", ids).
		Where("a.is_published = 1").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	res := r.batchMapArticles(ctx, rows)

	idOrder := make(map[uint64]int, len(ids))
	for i, id := range ids {
		idOrder[id] = i
	}

	sort.Slice(res, func(i, j int) bool {
		return idOrder[res[i].ID] < idOrder[res[j].ID]
	})

	return res, nil
}

func (r *Repository) LikeArticle(ctx context.Context, articleID uint64) (int, error) {
	var likeNum int
	err := r.db.WithContext(ctx).Model(&tArticle{}).
		Where("id = ?", articleID).
		UpdateColumn("like_num", gorm.Expr("like_num + 1")).
		Pluck("like_num", &likeNum).Error
	if err != nil {
		return 0, err
	}
	return likeNum, nil
}

func (r *Repository) RandomArticle(ctx context.Context, excludeID uint64) (*domarticle.Article, error) {
	var row tArticle
	query := r.db.WithContext(ctx).Where("is_published = 1")
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Order("rand()").First(&row).Error
	if err != nil {
		return nil, err
	}
	tags, _ := r.queryArticleTags(ctx, row.ID)
	return &domarticle.Article{
		ID:          row.ID,
		Title:       row.Title,
		Slug:        row.Slug,
		TitleImage:  row.TitleImage,
		Description: row.Description,
		ReadNum:     row.ReadNum,
		LikeNum:     row.LikeNum,
		IsTop:       row.IsTop,
		IsPublished: row.IsPublished,
		Tags:        tags,
		CreateTime:  row.CreateTime,
		UpdateTime:  row.UpdateTime,
	}, nil
}

// GetArticleBySlug 根据 slug 获取文章详情
func (r *Repository) GetArticleBySlug(ctx context.Context, slug string) (*domarticle.Article, error) {
	var row tArticle
	err := r.db.WithContext(ctx).Where("slug = ? AND is_published = 1", slug).First(&row).Error
	if err != nil {
		return nil, err
	}

	// 获取内容
	var content tArticleContent
	err = r.db.WithContext(ctx).Where("article_id = ?", row.ID).First(&content).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 获取分类
	var categoryID uint64
	r.db.WithContext(ctx).Table("t_article_category_rel").
		Where("article_id = ?", row.ID).
		Select("MIN(category_id)").Scan(&categoryID)

	var categoryName string
	if categoryID > 0 {
		r.db.WithContext(ctx).Table("t_category").
			Where("id = ?", categoryID).
			Select("name").Scan(&categoryName)
	}

	// 获取标签
	tags, _ := r.queryArticleTags(ctx, row.ID)

	// 获取专题
	topics, _ := r.queryArticleTopics(ctx, row.ID)

	return &domarticle.Article{
		ID:          row.ID,
		Title:       row.Title,
		Slug:        row.Slug,
		TitleImage:  row.TitleImage,
		Description: row.Description,
		Content:     content.Content,
		ReadNum:     row.ReadNum,
		ShareNum:    row.ShareNum,
		LikeNum:     row.LikeNum,
		IsTop:       row.IsTop,
		IsPublished: row.IsPublished,
		PublishTime: row.PublishTime,
		CategoryID:  categoryID,
		Category:    categoryName,
		Tags:        tags,
		Topics:      topics,
		CreateTime:  row.CreateTime,
		UpdateTime:  row.UpdateTime,
	}, nil
}

// FindScheduledArticles 查询待发布文章（is_published=0 且 publish_time <= before 且 publish_time IS NOT NULL）
func (r *Repository) FindScheduledArticles(ctx context.Context, before time.Time) ([]domarticle.Article, error) {
	var rows []tArticle
	err := r.db.WithContext(ctx).
		Where("is_published = 0 AND publish_time IS NOT NULL AND publish_time <= ?", before).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	articles := make([]domarticle.Article, 0, len(rows))
	for _, row := range rows {
		articles = append(articles, domarticle.Article{
			ID:          row.ID,
			Title:       row.Title,
			Slug:        row.Slug,
			IsPublished: row.IsPublished,
			PublishTime: row.PublishTime,
		})
	}
	return articles, nil
}

// PublishScheduledArticle 发布定时文章（设置 is_published=1，清空 publish_time）
func (r *Repository) PublishScheduledArticle(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&tArticle{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_published": 1,
			"publish_time": nil,
			"update_time":  time.Now(),
		}).Error
}
