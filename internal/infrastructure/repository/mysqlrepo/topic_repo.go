package mysqlrepo

import (
	"context"
	"errors"
	"sort"
	"time"

	domarticle "sanmoo-server-go/internal/domain/article"
	apperr "sanmoo-server-go/internal/shared/errors"

	"gorm.io/gorm"
)

// ------ Admin CRUD ------

func (r *Repository) ListAllWithCountTopics(ctx context.Context) ([]tTopic, []int64, error) {
	type row struct {
		ID           uint64
		Name         string
		Description  string
		CoverImage   string
		CreateTime   time.Time
		UpdateTime   time.Time
		ArticleCount int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("t_topic t").
		Select(`t.id, t.name, t.description, t.cover_image, t.create_time, t.update_time, count(distinct atr.article_id) as article_count`).
		Joins("left join t_article_topic_rel atr on atr.topic_id = t.id").
		Joins("left join t_article a on a.id = atr.article_id and a.is_published = 1").
		Where("t.status = 1").
		Group("t.id").
		Order("t.create_time desc").
		Scan(&rows).Error
	if err != nil {
		return nil, nil, err
	}
	topics := make([]tTopic, 0, len(rows))
	counts := make([]int64, 0, len(rows))
	for _, r := range rows {
		topics = append(topics, tTopic{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			CoverImage:  r.CoverImage,
			CreateTime:  r.CreateTime,
			UpdateTime:  r.UpdateTime,
		})
		counts = append(counts, r.ArticleCount)
	}
	return topics, counts, nil
}

func (r *Repository) FindByIDTopic(ctx context.Context, id uint64) (*tTopic, int64, error) {
	var out tTopic
	if err := r.db.WithContext(ctx).Table("t_topic").Where("id = ? and status = 1", id).Take(&out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperr.ErrNotFound
		}
		return nil, 0, err
	}
	var count int64
	if err := r.db.WithContext(ctx).Table("t_article_topic_rel atr").
		Joins("join t_article a on a.id = atr.article_id and a.is_published = 1").
		Where("atr.topic_id = ?", id).
		Count(&count).Error; err != nil {
		return nil, 0, err
	}
	return &out, count, nil
}

func (r *Repository) CreateTopic(ctx context.Context, name, description, coverImage string) (uint64, error) {
	// 先检查同名专题是否存在
	var existing tTopic
	err := r.db.WithContext(ctx).Table("t_topic").Where("name = ?", name).Take(&existing).Error
	if err == nil {
		// 同名专题已存在
		if existing.Status == 1 {
			return 0, apperr.New(apperr.ErrConflict.Code, "专题名称已存在")
		}
		// status=0 的专题，恢复并更新
		updates := map[string]any{
			"status":      1,
			"description": description,
			"cover_image": coverImage,
		}
		if res := r.db.WithContext(ctx).Table("t_topic").Where("id = ?", existing.ID).Updates(updates); res.Error != nil {
			return 0, res.Error
		}
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	// 不存在则新建
	t := tTopic{Name: name, Description: description, CoverImage: coverImage, Status: 1}
	if err := r.db.WithContext(ctx).Create(&t).Error; err != nil {
		return 0, err
	}
	return t.ID, nil
}

func (r *Repository) GetTopicArticleIDs(ctx context.Context, topicID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Table("t_article_topic_rel").
		Select("article_id").
		Where("topic_id = ?", topicID).
		Order("sort_order asc, id asc").
		Pluck("article_id", &ids).Error
	return ids, err
}

func (r *Repository) SetTopicArticles(ctx context.Context, topicID uint64, articleIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("t_article_topic_rel").Where("topic_id = ?", topicID).Delete(nil).Error; err != nil {
			return err
		}
		for i, aid := range articleIDs {
			rel := tArticleTopicRel{ArticleID: aid, TopicID: topicID, SortOrder: i + 1}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) ListPublishedArticleOptions(ctx context.Context) ([]struct {
	ID    uint64
	Title string
}, error) {
	type row struct {
		ID    uint64
		Title string
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("t_article").
		Select("id, title").
		Where("is_published = 1").
		Order("id desc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]struct {
		ID    uint64
		Title string
	}, len(rows))
	for i, r := range rows {
		out[i].ID = r.ID
		out[i].Title = r.Title
	}
	return out, nil
}

func (r *Repository) UpdateTopic(ctx context.Context, id uint64, name, description, coverImage string) error {
	// 先检查专题是否存在（不能依赖 RowsAffected，因为 MySQL 在值未变化时返回 0）
	var count int64
	if err := r.db.WithContext(ctx).Table("t_topic").Where("id = ? and status = 1", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return apperr.ErrNotFound
	}
	updates := map[string]any{"name": name, "description": description, "cover_image": coverImage}
	return r.db.WithContext(ctx).Table("t_topic").Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) DeleteTopic(ctx context.Context, id uint64) error {
	// 检查是否有关联文章
	var relCount int64
	if err := r.db.WithContext(ctx).Table("t_article_topic_rel").Where("topic_id = ?", id).Count(&relCount).Error; err != nil {
		return err
	}
	if relCount > 0 {
		return apperr.New(apperr.ErrConflict.Code, "专题正在被文章使用，无法删除")
	}
	res := r.db.WithContext(ctx).Table("t_topic").Where("id = ? and status = 1", id).Update("status", 0)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *Repository) BatchDeleteTopics(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	// 检查关联文章
	var relCount int64
	if err := r.db.WithContext(ctx).Table("t_article_topic_rel").Where("topic_id in ?", ids).Count(&relCount).Error; err != nil {
		return err
	}
	if relCount > 0 {
		// 找出具体哪些专题被使用
		var names []string
		if err := r.db.WithContext(ctx).Table("t_topic").Select("name").Where("id in ?", ids).Scan(&names).Error; err != nil {
			return err
		}
		if len(names) > 0 {
			return apperr.New(apperr.ErrConflict.Code, "以下专题正在被文章使用，无法删除："+stringsJoin(names, ", "))
		}
		return apperr.New(apperr.ErrConflict.Code, "专题正在被文章使用，无法删除")
	}
	return r.db.WithContext(ctx).Table("t_topic").Where("id in ? and status = 1", ids).Update("status", 0).Error
}

func stringsJoin(parts []string, sep string) string {
	result := ""
	for i, s := range parts {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// ------ 小程序端查询 ------

func (r *Repository) ListTopics(ctx context.Context, page, size int) ([]tTopic, []int64, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	var total int64
	if err := r.db.WithContext(ctx).Table("t_topic t").Where("t.status = 1").Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}
	type row struct {
		ID           uint64
		Name         string
		Description  string
		CoverImage   string
		CreateTime   time.Time
		ArticleCount int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("t_topic t").
		Select(`t.id, t.name, t.description, t.cover_image, t.create_time, count(distinct atr.article_id) as article_count`).
		Joins("left join t_article_topic_rel atr on atr.topic_id = t.id").
		Joins("left join t_article a on a.id = atr.article_id and a.is_published = 1").
		Where("t.status = 1").
		Group("t.id").
		Order("t.create_time desc").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error
	if err != nil {
		return nil, nil, 0, err
	}
	topics := make([]tTopic, 0, len(rows))
	counts := make([]int64, 0, len(rows))
	for _, r := range rows {
		topics = append(topics, tTopic{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			CoverImage:  r.CoverImage,
			CreateTime:  r.CreateTime,
		})
		counts = append(counts, r.ArticleCount)
	}
	return topics, counts, total, nil
}

func (r *Repository) GetTopicByID(ctx context.Context, id uint64) (tTopic, int64, error) {
	var out tTopic
	if err := r.db.WithContext(ctx).Table("t_topic").Where("id = ? and status = 1", id).Take(&out).Error; err != nil {
		return tTopic{}, 0, err
	}
	var count int64
	err := r.db.WithContext(ctx).Table("t_article_topic_rel atr").
		Joins("join t_article a on a.id = atr.article_id and a.is_published = 1").
		Where("atr.topic_id = ?", id).
		Count(&count).Error
	if err != nil {
		return tTopic{}, 0, err
	}
	return out, count, nil
}

func (r *Repository) ListArticlesByTopic(ctx context.Context, topicID uint64, page, size int, isPublished *bool) ([]domarticle.Article, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	base := r.db.WithContext(ctx).Table("t_article a").
		Joins("join t_article_topic_rel atr on atr.article_id = a.id").
		Where("atr.topic_id = ?", topicID)
	if isPublished != nil {
		base = base.Where("a.is_published = ?", *isPublished)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
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
	err := base.
		Select("a.id, a.title, a.title_image, a.description, a.read_num, a.is_top, a.is_published, a.create_time, a.update_time, acr.category_id, c.name as category_name").
		Joins("left join (select article_id, min(category_id) as category_id from t_article_category_rel group by article_id) acr on acr.article_id = a.id").
		Joins("left join t_category c on c.id = acr.category_id").
		Order("atr.sort_order asc, a.create_time desc").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	res := make([]domarticle.Article, 0, len(rows))
	for _, row := range rows {
		tags, _ := r.queryArticleTags(ctx, row.ID)
		res = append(res, domarticle.Article{ID: row.ID, Title: row.Title, TitleImage: row.TitleImage, Description: row.Description, ReadNum: row.ReadNum, IsTop: row.IsTop, IsPublished: row.IsPublished, CategoryID: row.CategoryID, Category: row.CategoryName, Tags: tags, CreateTime: row.CreateTime, UpdateTime: row.UpdateTime})
	}
	return res, total, nil
}

func (r *Repository) queryArticleTags(ctx context.Context, articleID uint64) ([]domarticle.TagRef, error) {
	type tr struct {
		ID   uint64
		Name string
	}
	var rows []tr
	err := r.db.WithContext(ctx).Table("t_article_tag_rel atr").Select("t.id, t.name").Joins("join t_tag t on t.id = atr.tag_id").Where("atr.article_id = ?", articleID).Order("t.id asc").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	tags := make([]domarticle.TagRef, 0, len(rows))
	for _, t := range rows {
		tags = append(tags, domarticle.TagRef{ID: t.ID, Name: t.Name})
	}
	return tags, nil
}

func (r *Repository) queryArticleTopics(ctx context.Context, articleID uint64) ([]domarticle.TopicRef, error) {
	type tr struct {
		ID   uint64
		Name string
	}
	var rows []tr
	err := r.db.WithContext(ctx).Table("t_article_topic_rel atr").Select("t.id, t.name").Joins("join t_topic t on t.id = atr.topic_id").Where("atr.article_id = ?", articleID).Order("atr.sort_order asc").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	topics := make([]domarticle.TopicRef, 0, len(rows))
	for _, t := range rows {
		topics = append(topics, domarticle.TopicRef{ID: t.ID, Name: t.Name})
	}
	return topics, nil
}

// batchQueryArticleTags 批量查询文章的标签，返回 articleID → []TagRef 的映射，避免 N+1 查询。
func (r *Repository) batchQueryArticleTags(ctx context.Context, articleIDs []uint64) map[uint64][]domarticle.TagRef {
	if len(articleIDs) == 0 {
		return make(map[uint64][]domarticle.TagRef)
	}
	type tr struct {
		ArticleID uint64
		TagID     uint64
		TagName   string
	}
	var rows []tr
	err := r.db.WithContext(ctx).Table("t_article_tag_rel atr").
		Select("atr.article_id, t.id as tag_id, t.name as tag_name").
		Joins("join t_tag t on t.id = atr.tag_id").
		Where("atr.article_id in (?)", articleIDs).
		Order("t.id asc").
		Scan(&rows).Error
	if err != nil {
		return make(map[uint64][]domarticle.TagRef)
	}
	result := make(map[uint64][]domarticle.TagRef, len(articleIDs))
	for _, row := range rows {
		result[row.ArticleID] = append(result[row.ArticleID], domarticle.TagRef{ID: row.TagID, Name: row.TagName})
	}
	return result
}

func (r *Repository) FindByIDArticle(ctx context.Context, id uint64) (*domarticle.Article, error) {
	type row struct {
		ID                                                    uint64
		Title, TitleImage, Description, Content, CategoryName string
		CategoryID                                            uint64
		ReadNum, LikeNum                                      int
		IsTop, IsPublished                                    bool
		CreateTime, UpdateTime                                time.Time
	}
	var out row
	err := r.db.WithContext(ctx).Table("t_article a").
		Select("a.id, a.title, a.title_image, a.description, ac.content, a.read_num, a.like_num, a.is_top, a.is_published, a.create_time, a.update_time, acr.category_id, c.name as category_name").
		Joins("left join t_article_content ac on ac.article_id = a.id").
		Joins("left join t_article_category_rel acr on acr.article_id = a.id").
		Joins("left join t_category c on c.id = acr.category_id").
		Where("a.id = ?", id).Limit(1).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	if out.ID == 0 {
		return nil, apperr.ErrNotFound
	}
	tags, _ := r.queryArticleTags(ctx, id)
	topics, _ := r.queryArticleTopics(ctx, id)
	return &domarticle.Article{ID: out.ID, Title: out.Title, TitleImage: out.TitleImage, Description: out.Description, Content: out.Content, ReadNum: out.ReadNum, LikeNum: out.LikeNum, IsTop: out.IsTop, IsPublished: out.IsPublished, CategoryID: out.CategoryID, Category: out.CategoryName, Tags: tags, Topics: topics, CreateTime: out.CreateTime, UpdateTime: out.UpdateTime}, nil
}

func (r *Repository) FindPrev(ctx context.Context, id uint64) (*domarticle.Article, error) {
	var row tArticle
	if err := r.db.WithContext(ctx).Where("id < ? and is_published=1", id).Order("id desc").First(&row).Error; err != nil {
		return nil, err
	}
	return &domarticle.Article{ID: row.ID, Title: row.Title}, nil
}

func (r *Repository) FindNext(ctx context.Context, id uint64) (*domarticle.Article, error) {
	var row tArticle
	if err := r.db.WithContext(ctx).Where("id > ? and is_published=1", id).Order("id asc").First(&row).Error; err != nil {
		return nil, err
	}
	return &domarticle.Article{ID: row.ID, Title: row.Title}, nil
}

func (r *Repository) CreateArticle(ctx context.Context, a *domarticle.Article) (uint64, error) {
	var id uint64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		am := tArticle{Title: a.Title, TitleImage: a.TitleImage, Description: a.Description, ReadNum: 1, IsTop: a.IsTop, IsPublished: a.IsPublished}
		if err := tx.Create(&am).Error; err != nil {
			return err
		}
		id = am.ID
		if err := tx.Create(&tArticleContent{ArticleID: id, Content: a.Content}).Error; err != nil {
			return err
		}
		if err := tx.Create(&tArticleCategoryRel{ArticleID: id, CategoryID: a.CategoryID}).Error; err != nil {
			return err
		}
		for _, tg := range a.Tags {
			if err := tx.Create(&tArticleTagRel{ArticleID: id, TagID: tg.ID}).Error; err != nil {
				return err
			}
		}
		for i, tp := range a.Topics {
			if err := tx.Create(&tArticleTopicRel{ArticleID: id, TopicID: tp.ID, SortOrder: i + 1}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return id, err
}

func (r *Repository) UpdateArticleStatus(ctx context.Context, id uint64, isPublished, isTop *bool) error {
	updates := map[string]any{}
	if isPublished != nil {
		updates["is_published"] = *isPublished
	}
	if isTop != nil {
		updates["is_top"] = *isTop
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&tArticle{}).Where("id=?", id).Updates(updates).Error
}

func (r *Repository) BatchUpdateArticleStatus(ctx context.Context, ids []uint64, isPublished, isTop *bool) error {
	updates := map[string]any{}
	if isPublished != nil {
		updates["is_published"] = *isPublished
	}
	if isTop != nil {
		updates["is_top"] = *isTop
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&tArticle{}).Where("id in ?", ids).Updates(updates).Error
}

func (r *Repository) UpdateArticle(ctx context.Context, a *domarticle.Article) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&tArticle{}).Where("id=?", a.ID).Updates(map[string]any{"title": a.Title, "title_image": a.TitleImage, "description": a.Description, "is_top": a.IsTop, "is_published": a.IsPublished}).Error; err != nil {
			return err
		}
		if err := tx.Model(&tArticleContent{}).Where("article_id=?", a.ID).Update("content", a.Content).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id=?", a.ID).Delete(&tArticleCategoryRel{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&tArticleCategoryRel{ArticleID: a.ID, CategoryID: a.CategoryID}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id=?", a.ID).Delete(&tArticleTagRel{}).Error; err != nil {
			return err
		}
		for _, tg := range a.Tags {
			if err := tx.Create(&tArticleTagRel{ArticleID: a.ID, TagID: tg.ID}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("article_id=?", a.ID).Delete(&tArticleTopicRel{}).Error; err != nil {
			return err
		}
		for i, tp := range a.Topics {
			if err := tx.Create(&tArticleTopicRel{ArticleID: a.ID, TopicID: tp.ID, SortOrder: i + 1}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *Repository) DeleteArticle(ctx context.Context, id uint64) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Where("article_id = ?", id).Delete(&tArticleContent{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("article_id = ?", id).Delete(&tArticleCategoryRel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("article_id = ?", id).Delete(&tArticleTagRel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("article_id = ?", id).Delete(&tArticleTopicRel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&tArticle{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
func (r *Repository) BatchDeleteArticle(ctx context.Context, ids []uint64) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Where("article_id in ?", ids).Delete(&tArticleContent{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("article_id in ?", ids).Delete(&tArticleCategoryRel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("article_id in ?", ids).Delete(&tArticleTagRel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("id in ?", ids).Delete(&tArticle{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *Repository) ArchiveList(ctx context.Context) ([]domarticle.ArchiveItem, error) {
	var rows []tArticle
	if err := r.db.WithContext(ctx).Where("is_published=1").Order("create_time desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	group := map[string][]domarticle.Article{}
	for _, row := range rows {
		month := row.CreateTime.Format("2006-01")
		group[month] = append(group[month], domarticle.Article{ID: row.ID, Title: row.Title, CreateTime: row.CreateTime})
	}
	months := make([]string, 0, len(group))
	for k := range group {
		months = append(months, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(months)))
	out := make([]domarticle.ArchiveItem, 0, len(months))
	for _, m := range months {
		out = append(out, domarticle.ArchiveItem{Month: m, Items: group[m]})
	}
	return out, nil
}

func (r *Repository) IncreaseReadAndPV(ctx context.Context, articleID uint64) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&tArticle{}).Where("id=?", articleID).UpdateColumn("read_num", gorm.Expr("read_num + 1")).Error; err != nil {
			return err
		}
		var row tPV
		err := tx.Where("article_id=? and pv_date=?", articleID, today).First(&row).Error
		if err == nil {
			return tx.Model(&tPV{}).Where("article_id=? and pv_date=?", articleID, today).UpdateColumn("pv_count", gorm.Expr("pv_count + 1")).Error
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&tPV{ArticleID: articleID, PVDate: today, PVCount: 1}).Error
		}
		return err
	})
}

func (r *Repository) IncreaseShareCount(ctx context.Context, articleID uint64) error {
	return r.db.WithContext(ctx).Model(&tArticle{}).Where("id=?", articleID).UpdateColumn("share_num", gorm.Expr("share_num + 1")).Error
}

func (r *Repository) PVStatistics(ctx context.Context, days int) ([]domarticle.PVPoint, error) {
	if days <= 0 {
		days = 7
	}
	from := time.Now().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)
	type row struct {
		PVDate  time.Time
		PVCount int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("t_statistics_article_pv").Select("pv_date, sum(pv_count) as pv_count").Where("pv_date >= ?", from).Group("pv_date").Order("pv_date asc").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	res := make([]domarticle.PVPoint, 0, len(rows))
	for _, row := range rows {
		res = append(res, domarticle.PVPoint{Date: row.PVDate.Format("2006-01-02"), Count: row.PVCount})
	}
	return res, nil
}

func (r *Repository) RecordSearchHistory(ctx context.Context, keyword string) error {
	return r.db.WithContext(ctx).Create(&tSearchHistory{
		Keyword:    keyword,
		SearchTime: time.Now(),
	}).Error
}

func (r *Repository) GetHotSearchKeywords(ctx context.Context, limit int) ([]string, error) {
	type row struct {
		Keyword string
		Count   int64
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("t_search_history").
		Select("keyword, count(*) as count").
		Where("search_time >= ?", time.Now().AddDate(0, 0, -7)).
		Group("keyword").
		Order("count desc").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.Keyword)
	}
	return result, nil
}
