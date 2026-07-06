package mysqlrepo

import (
	"context"
	"math"
	"strings"
	"time"

	dommpuser "sanmoo-server-go/internal/domain/mpuser"

	"gorm.io/gorm"
)

// ======================== MP User Admin Repository ========================

// ListMPUsers 分页查询微信用户列表（含标签数、浏览数、收藏数统计）。
func (r *Repository) ListMPUsers(ctx context.Context, q dommpuser.ListQuery) ([]dommpuser.MPUserSummary, int64, error) {
	var total int64
	db := r.db.WithContext(ctx).Table("t_mp_user")

	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		db = db.Where("nickname LIKE ? OR openid LIKE ?", kw, kw)
	}
	if q.TagName != "" {
		db = db.Where("EXISTS (SELECT 1 FROM t_mp_user_tag WHERE openid = t_mp_user.openid AND tag_name = ?)", q.TagName)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 子查询：标签数、浏览数、收藏数
	type row struct {
		ID             uint64    `gorm:"column:id"`
		OpenID         string    `gorm:"column:openid"`
		Nickname       string    `gorm:"column:nickname"`
		Avatar         string    `gorm:"column:avatar"`
		Status         uint8     `gorm:"column:status"`
		FirstLoginTime time.Time `gorm:"column:first_login_time"`
		LastLoginTime  time.Time `gorm:"column:last_login_time"`
		CreateTime     time.Time `gorm:"column:create_time"`
		TagCount       int64     `gorm:"column:tag_count"`
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
		COALESCE((SELECT COUNT(1) FROM t_mp_user_tag WHERE openid = t_mp_user.openid), 0) AS tag_count,
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
			TagCount:       int(row.TagCount),
			ViewCount:      row.ViewCount,
			FavoriteCount:  row.FavoriteCount,
			CreateTime:     row.CreateTime,
		})
	}
	return out, total, nil
}

// GetMPUserDetail 获取微信用户详情（含标签、兴趣、画像）。
func (r *Repository) GetMPUserDetail(ctx context.Context, openID string) (*dommpuser.MPUserDetail, error) {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var u tMPUser
	if err := r.db.WithContext(ctx).Where("openid = ?", openID).Take(&u).Error; err != nil {
		return nil, err
	}

	viewCount, favoriteCount, totalStay, _ := r.CountMPUserBehavior(ctx, openID)
	tags, _ := r.GetMPUserTags(ctx, openID)
	interests, _ := r.GetMPUserInterests(ctx, openID)
	profile, _ := r.GetMPUserProfile(ctx, openID)

	return &dommpuser.MPUserDetail{
		ID:               u.ID,
		OpenID:           u.OpenID,
		Nickname:         u.Nickname,
		Avatar:           u.Avatar,
		Status:           u.Status,
		FirstLoginTime:   u.FirstLoginTime,
		LastLoginTime:    u.LastLoginTime,
		ViewCount:        viewCount,
		FavoriteCount:    favoriteCount,
		TotalStaySeconds: totalStay,
		Tags:             tags,
		Interests:        interests,
		Profile:          profile,
		CreateTime:       u.CreateTime,
	}, nil
}

// GetMPUserTags 获取用户标签列表。
func (r *Repository) GetMPUserTags(ctx context.Context, openID string) ([]dommpuser.UserTag, error) {
	var rows []tMPUserTag
	if err := r.db.WithContext(ctx).Where("openid = ?", strings.TrimSpace(openID)).
		Order("score DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dommpuser.UserTag, 0, len(rows))
	for _, row := range rows {
		out = append(out, dommpuser.UserTag{
			ID:          row.ID,
			TagName:     row.TagName,
			TagCategory: row.TagCategory,
			Score:       row.Score,
			Source:      row.Source,
			CreateTime:  row.CreateTime,
		})
	}
	return out, nil
}

// UpsertMPUserTag 新增或更新用户标签。
func (r *Repository) UpsertMPUserTag(ctx context.Context, openID, tagName, tagCategory string, score float64, source string) error {
	return r.db.WithContext(ctx).Exec(`
INSERT INTO t_mp_user_tag(openid, tag_name, tag_category, score, source)
VALUES(?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE score = VALUES(score), tag_category = VALUES(tag_category), source = VALUES(source), update_time = CURRENT_TIMESTAMP
`, strings.TrimSpace(openID), strings.TrimSpace(tagName), strings.TrimSpace(tagCategory), score, strings.TrimSpace(source)).Error
}

// DeleteMPUserTag 删除用户标签。
func (r *Repository) DeleteMPUserTag(ctx context.Context, tagID uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", tagID).Delete(&tMPUserTag{}).Error
}

// GetMPUserProfile 获取用户六边形画像。
func (r *Repository) GetMPUserProfile(ctx context.Context, openID string) (*dommpuser.UserProfile, error) {
	var rows []tMPUserProfile
	if err := r.db.WithContext(ctx).Where("openid = ?", strings.TrimSpace(openID)).
		Order("dimension").Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	dims := make([]dommpuser.ProfileDimension, 0, len(rows))
	var latest *time.Time
	for _, row := range rows {
		dims = append(dims, dommpuser.ProfileDimension{
			Dimension: row.Dimension,
			Score:     row.Score,
		})
		if latest == nil || row.UpdateTime.After(*latest) {
			t := row.UpdateTime
			latest = &t
		}
	}
	return &dommpuser.UserProfile{Dimensions: dims, UpdatedAt: latest}, nil
}

// SaveMPUserProfile 保存单个画像维度得分。
func (r *Repository) SaveMPUserProfile(ctx context.Context, openID, dimension string, score float64) error {
	return r.db.WithContext(ctx).Exec(`
INSERT INTO t_mp_user_profile(openid, dimension, score)
VALUES(?, ?, ?)
ON DUPLICATE KEY UPDATE score = VALUES(score), update_time = CURRENT_TIMESTAMP
`, strings.TrimSpace(openID), strings.TrimSpace(dimension), score).Error
}

// ComputeAndSaveProfile 基于行为数据计算并保存六边形画像。
// 六个维度：技术深度、活跃度、阅读广度、收藏倾向、互动程度、学习投入。
func (r *Repository) ComputeAndSaveProfile(ctx context.Context, openID string) (*dommpuser.UserProfile, error) {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	// 1. 活跃度：综合最近30天浏览历史 + 行为记录
	var recentBehaviorCount int64
	r.db.WithContext(ctx).Table("t_mp_browse_history").
		Where("openid = ? AND update_time >= DATE_SUB(NOW(), INTERVAL 30 DAY)", openID).
		Count(&recentBehaviorCount)

	// 2. 阅读广度：基于浏览历史 + 行为记录中的不同文章数
	var distinctArticles int64
	r.db.WithContext(ctx).Table("t_mp_browse_history").
		Where("openid = ?", openID).
		Distinct("article_id").Count(&distinctArticles)
	if distinctArticles == 0 {
		r.db.WithContext(ctx).Table("t_mp_user_behavior").
			Where("openid = ? AND event_type = 'view'", openID).
			Distinct("article_id").Count(&distinctArticles)
	}

	// 3. 收藏倾向：收藏数 / (收藏数 + 浏览数) * 100
	var favCount, viewCount int64
	r.db.WithContext(ctx).Table("t_mp_user_favorite").Where("openid = ?", openID).Count(&favCount)
	r.db.WithContext(ctx).Table("t_mp_browse_history").Where("openid = ?", openID).Count(&viewCount)
	if viewCount == 0 {
		r.db.WithContext(ctx).Table("t_mp_user_behavior").Where("openid = ? AND event_type = 'view'", openID).Count(&viewCount)
	}

	var favRatio float64
	if viewCount > 0 {
		favRatio = math.Min(float64(favCount)/float64(viewCount)*100, 100)
	}

	// 4. 技术深度：基于兴趣分数（tag/category维度的平均分归一化）
	var avgInterestScore float64
	r.db.WithContext(ctx).Table("t_mp_user_interest").
		Where("openid = ?", openID).
		Select("COALESCE(AVG(score), 0)").Scan(&avgInterestScore)
	techDepth := math.Min(avgInterestScore/200*100, 100)

	// 5. 学习投入：基于总停留时长（秒），归一化到0-100
	var totalStay int64
	r.db.WithContext(ctx).Table("t_mp_user_behavior").
		Where("openid = ? AND event_type = 'stay'", openID).
		Select("COALESCE(SUM(stay_seconds), 0)").Scan(&totalStay)
	learningScore := math.Min(float64(totalStay)/3600*100, 100) // 假设 3600秒(1小时) = 满分100

	// 6. 互动程度：基于不同行为类型数 (view/stay/favorite)
	var behaviorTypes int64
	r.db.WithContext(ctx).Table("t_mp_user_behavior").
		Where("openid = ?", openID).
		Distinct("event_type").Count(&behaviorTypes)
	interactionScore := math.Min(float64(behaviorTypes)/5*100, 100) // 5种类型 = 满分

	// 活跃度归一化（假设30天内50次行为 = 满分100）
	activityScore := math.Min(float64(recentBehaviorCount)/50*100, 100)

	// 阅读广度归一化（假设看了30篇不同文章 = 满分100）
	breadthScore := math.Min(float64(distinctArticles)/30*100, 100)

	dims := map[string]float64{
		"活跃度":  math.Round(activityScore*10) / 10,
		"阅读广度": math.Round(breadthScore*10) / 10,
		"收藏倾向": math.Round(favRatio*10) / 10,
		"技术深度": math.Round(techDepth*10) / 10,
		"学习投入": math.Round(learningScore*10) / 10,
		"互动程度": math.Round(interactionScore*10) / 10,
	}

	for dim, score := range dims {
		if err := r.SaveMPUserProfile(ctx, openID, dim, score); err != nil {
			return nil, err
		}
	}

	return r.GetMPUserProfile(ctx, openID)
}

// ComputeAndSaveTags 基于行为数据自动生成行为标签和兴趣维度。
func (r *Repository) ComputeAndSaveTags(ctx context.Context, openID string) ([]dommpuser.UserTag, error) {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	// 清除旧的自动行为标签
	r.db.WithContext(ctx).Where("openid = ? AND source = 'auto'", openID).Delete(&tMPUserTag{})

	// 清除旧的兴趣维度（tag 类型）
	r.db.WithContext(ctx).Where("openid = ? AND dimension_type = 'tag'", openID).Delete(&tMPUserInterest{})

	// 1. 计算行为标签（写入 t_mp_user_tag）
	r.computeBehaviorTags(ctx, openID)

	// 2. 计算兴趣标签（写入 t_mp_user_interest）
	r.computeInterestTags(ctx, openID)

	return r.GetMPUserTags(ctx, openID)
}

// computeBehaviorTags 基于用户行为指标计算行为标签。
func (r *Repository) computeBehaviorTags(ctx context.Context, openID string) {
	var viewCount, favCount, stayTotal int64
	var recentBehaviorCount int64
	var firstLoginTime time.Time

	// 浏览次数优先从浏览历史表统计
	r.db.WithContext(ctx).Table("t_mp_browse_history").Where("openid = ?", openID).Count(&viewCount)
	if viewCount == 0 {
		r.db.WithContext(ctx).Table("t_mp_user_behavior").Where("openid = ? AND event_type = 'view'", openID).Count(&viewCount)
	}
	r.db.WithContext(ctx).Table("t_mp_user_favorite").Where("openid = ?", openID).Count(&favCount)
	r.db.WithContext(ctx).Table("t_mp_user_behavior").Where("openid = ? AND event_type = 'stay'", openID).Select("COALESCE(SUM(stay_seconds), 0)").Scan(&stayTotal)
	// 近期行为：综合浏览历史和收藏记录
	r.db.WithContext(ctx).Table("t_mp_browse_history").
		Where("openid = ? AND update_time >= DATE_SUB(NOW(), INTERVAL 30 DAY)", openID).
		Count(&recentBehaviorCount)

	var u tMPUser
	if err := r.db.WithContext(ctx).Where("openid = ?", openID).Take(&u).Error; err == nil {
		firstLoginTime = u.FirstLoginTime
	}

	// 浏览等级
	if viewCount >= 50 {
		_ = r.UpsertMPUserTag(ctx, openID, "重度浏览者", "behavior", 90, "auto")
	} else if viewCount >= 20 {
		_ = r.UpsertMPUserTag(ctx, openID, "中度浏览者", "behavior", 60, "auto")
	} else if viewCount > 0 {
		_ = r.UpsertMPUserTag(ctx, openID, "轻度浏览者", "behavior", 30, "auto")
	}

	// 收藏倾向
	if viewCount > 0 {
		favRatio := float64(favCount) / float64(viewCount)
		if favRatio >= 0.3 {
			_ = r.UpsertMPUserTag(ctx, openID, "收藏达人", "behavior", 90, "auto")
		} else if favRatio >= 0.1 {
			_ = r.UpsertMPUserTag(ctx, openID, "有收藏习惯", "behavior", 60, "auto")
		}
	}

	// 阅读深度
	if stayTotal >= 3600 {
		_ = r.UpsertMPUserTag(ctx, openID, "深度阅读者", "behavior", 90, "auto")
	} else if stayTotal >= 600 {
		_ = r.UpsertMPUserTag(ctx, openID, "中度阅读者", "behavior", 60, "auto")
	} else if stayTotal > 0 {
		_ = r.UpsertMPUserTag(ctx, openID, "轻度阅读者", "behavior", 30, "auto")
	}

	// 活跃度
	if recentBehaviorCount >= 30 {
		_ = r.UpsertMPUserTag(ctx, openID, "活跃用户", "behavior", 90, "auto")
	} else if recentBehaviorCount >= 10 {
		_ = r.UpsertMPUserTag(ctx, openID, "稳定用户", "behavior", 60, "auto")
	} else if recentBehaviorCount > 0 {
		_ = r.UpsertMPUserTag(ctx, openID, "低频用户", "behavior", 30, "auto")
	} else {
		_ = r.UpsertMPUserTag(ctx, openID, "沉默用户", "behavior", 10, "auto")
	}

	// 新老用户
	if !firstLoginTime.IsZero() && time.Since(firstLoginTime) <= 7*24*time.Hour {
		_ = r.UpsertMPUserTag(ctx, openID, "新用户", "behavior", 80, "auto")
	} else if !firstLoginTime.IsZero() && time.Since(firstLoginTime) >= 90*24*time.Hour {
		_ = r.UpsertMPUserTag(ctx, openID, "老用户", "behavior", 70, "auto")
	}
}

// computeInterestTags 基于用户阅读文章的真实标签计算兴趣维度（写入 t_mp_user_interest）。
func (r *Repository) computeInterestTags(ctx context.Context, openID string) {
	// 查询用户有过行为的文章及其行为类型（综合浏览历史 + 行为表）
	type articleBehavior struct {
		ArticleID uint64
		EventType string
	}
	var behaviors []articleBehavior

	// 从浏览历史表获取
	r.db.WithContext(ctx).Table("t_mp_browse_history").
		Where("openid = ?", openID).
		Select("article_id, 'view' as event_type").
		Find(&behaviors)

	// 从行为表补充（去重）
	existingIDs := make(map[uint64]bool)
	for _, b := range behaviors {
		existingIDs[b.ArticleID] = true
	}
	var behaviorRecords []articleBehavior
	r.db.WithContext(ctx).Table("t_mp_user_behavior").
		Where("openid = ?", openID).
		Select("article_id, event_type").
		Find(&behaviorRecords)
	for _, b := range behaviorRecords {
		if b.EventType != "view" || !existingIDs[b.ArticleID] {
			behaviors = append(behaviors, b)
			existingIDs[b.ArticleID] = true
		}
	}

	if len(behaviors) == 0 {
		return
	}

	// 构建 articleID → 加权总分 的映射
	// view: 1.0, stay: 2.0, favorite: 3.0, share: 2.0, like: 1.5
	weightMap := map[string]float64{
		"view":     1.0,
		"stay":     2.0,
		"favorite": 3.0,
		"share":    2.0,
		"like":     1.5,
		"click":    1.0,
	}
	articleScore := make(map[uint64]float64)
	for _, b := range behaviors {
		w, ok := weightMap[b.EventType]
		if !ok {
			w = 1.0
		}
		articleScore[b.ArticleID] += w
	}

	// 收集所有 articleID
	articleIDs := make([]uint64, 0, len(articleScore))
	for id := range articleScore {
		articleIDs = append(articleIDs, id)
	}

	// 查询这些文章关联的标签
	type articleTag struct {
		ArticleID uint64
		TagID     uint64
		TagName   string
	}
	var articleTags []articleTag
	r.db.WithContext(ctx).Table("t_article_tag_rel atr").
		Select("atr.article_id, t.id as tag_id, t.name as tag_name").
		Joins("JOIN t_tag t ON t.id = atr.tag_id").
		Where("atr.article_id IN (?)", articleIDs).
		Scan(&articleTags)

		// 按标签聚合加权分数（key: tag_id）
	type tagAgg struct {
		tagID uint64
		score float64
	}
	tagScoreMap := make(map[uint64]*tagAgg)
	for _, at := range articleTags {
		if existing, ok := tagScoreMap[at.TagID]; ok {
			existing.score += articleScore[at.ArticleID]
		} else {
			tagScoreMap[at.TagID] = &tagAgg{tagID: at.TagID, score: articleScore[at.ArticleID]}
		}
	}

	if len(tagScoreMap) == 0 {
		return
	}

	// 找到最大分用于归一化
	maxScore := 0.0
	for _, ag := range tagScoreMap {
		if ag.score > maxScore {
			maxScore = ag.score
		}
	}

	// 归一化到 0-100，取 Top 8
	type tagEntry struct {
		tagID uint64
		score float64
	}
	entries := make([]tagEntry, 0, len(tagScoreMap))
	for _, ag := range tagScoreMap {
		normalized := math.Round(ag.score/maxScore*100*10) / 10
		if normalized < 10 {
			normalized = 10
		}
		entries = append(entries, tagEntry{tagID: ag.tagID, score: normalized})
	}

	// 按分数降序排序
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].score > entries[i].score {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// 取 Top 8，写入 t_mp_user_interest
	limit := 8
	if len(entries) < limit {
		limit = len(entries)
	}
	now := time.Now()
	for i := 0; i < limit; i++ {
		interest := tMPUserInterest{
			OpenID:        openID,
			DimensionType: "tag",
			DimensionID:   entries[i].tagID,
			Score:         entries[i].score,
			UpdateTime:    now,
		}
		r.db.WithContext(ctx).Create(&interest)
	}
}

// ComputeAndSaveRadar 刷新雷达图数据：行为标签 + 兴趣维度 + 六边形画像。
func (r *Repository) ComputeAndSaveRadar(ctx context.Context, openID string) (*dommpuser.RadarData, error) {
	// 1. 行为标签 + 兴趣维度
	behaviorTags, err := r.ComputeAndSaveTags(ctx, openID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 2. 六边形画像
	profile, err := r.ComputeAndSaveProfile(ctx, openID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 3. 兴趣维度（含名称）
	interests, _ := r.GetMPUserInterests(ctx, openID)

	return &dommpuser.RadarData{
		Tags:      behaviorTags,
		Interests: interests,
		Profile:   profile,
	}, nil
}

// GetMPUserInterests 获取用户兴趣维度（含名称）。
func (r *Repository) GetMPUserInterests(ctx context.Context, openID string) ([]dommpuser.UserInterest, error) {
	var rows []tMPUserInterest
	if err := r.db.WithContext(ctx).Where("openid = ?", strings.TrimSpace(openID)).
		Order("score DESC").Limit(20).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dommpuser.UserInterest, 0, len(rows))
	for _, row := range rows {
		dimName := ""
		if row.DimensionType == "tag" {
			var tag tTag
			if err := r.db.WithContext(ctx).Where("id = ?", row.DimensionID).Take(&tag).Error; err == nil {
				dimName = tag.Name
			}
		} else if row.DimensionType == "category" {
			var cat tCategory
			if err := r.db.WithContext(ctx).Where("id = ?", row.DimensionID).Take(&cat).Error; err == nil {
				dimName = cat.Name
			}
		}
		out = append(out, dommpuser.UserInterest{
			DimensionType: row.DimensionType,
			DimensionID:   row.DimensionID,
			DimensionName: dimName,
			Score:         row.Score,
		})
	}
	return out, nil
}

// CountMPUserBehavior 统计用户行为次数。
func (r *Repository) CountMPUserBehavior(ctx context.Context, openID string) (viewCount int64, favoriteCount int64, totalStaySeconds int64, err error) {
	openID = strings.TrimSpace(openID)

	// 浏览次数优先从浏览历史表统计（数据更可靠）
	if err = r.db.WithContext(ctx).Table("t_mp_browse_history").
		Where("openid = ?", openID).Count(&viewCount).Error; err != nil {
		return
	}
	// 若浏览历史为空，回退到行为表
	if viewCount == 0 {
		if err = r.db.WithContext(ctx).Table("t_mp_user_behavior").
			Where("openid = ? AND event_type = 'view'", openID).Count(&viewCount).Error; err != nil {
			return
		}
	}
	if err = r.db.WithContext(ctx).Table("t_mp_user_favorite").
		Where("openid = ?", openID).Count(&favoriteCount).Error; err != nil {
		return
	}
	if err = r.db.WithContext(ctx).Table("t_mp_user_behavior").
		Where("openid = ? AND event_type = 'stay'", openID).
		Select("COALESCE(SUM(stay_seconds), 0)").Scan(&totalStaySeconds).Error; err != nil {
		return
	}
	return
}
