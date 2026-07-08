package article

import (
	"context"
	"fmt"
	"sort"

	"sanmoo-server-go/internal/application/recommendation"
	domarticle "sanmoo-server-go/internal/domain/article"
	domsetting "sanmoo-server-go/internal/domain/setting"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/infrastructure/logger"
	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/shared/pagination"
)

type Service struct {
	repo        domarticle.Repository
	cache       *cache.BusinessCache
	recRegistry *recommendation.Registry
	settingRepo domsetting.Repository
}

func NewService(repo domarticle.Repository, cache *cache.BusinessCache, recRegistry *recommendation.Registry, settingRepo domsetting.Repository) *Service {
	return &Service{repo: repo, cache: cache, recRegistry: recRegistry, settingRepo: settingRepo}
}

func (s *Service) ListArticles(ctx context.Context, page, size int, keyword string, categoryID, tagID uint64, isPublished *int) (*pagination.PageData, error) {
	var pub *bool
	var pubStr string
	if isPublished != nil {
		b := *isPublished == 1
		pub = &b
		pubStr = fmt.Sprintf("%d", *isPublished)
	} else {
		pubStr = "nil"
	}

	if s.cache != nil && keyword == "" {
		cacheKey := cache.BuildArticleListKey(page, size, keyword, categoryID, tagID, pubStr)
		var result pagination.PageData
		err := s.cache.GetOrSet(ctx, cacheKey, &result, func() (interface{}, error) {
			return s.queryList(ctx, page, size, keyword, categoryID, tagID, pub)
		}, cache.ShortTTLMin)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}

	return s.queryList(ctx, page, size, keyword, categoryID, tagID, pub)
}

func (s *Service) queryList(ctx context.Context, page, size int, keyword string, categoryID, tagID uint64, pub *bool) (*pagination.PageData, error) {
	q := domarticle.ListQuery{
		Page:        page,
		Size:        size,
		Keyword:     keyword,
		CategoryID:  categoryID,
		TagID:       tagID,
		IsPublished: pub,
	}
	list, total, err := s.repo.ListArticles(ctx, q)
	if err != nil {
		return nil, err
	}
	return pagination.NewPageData(list, total, page, size), nil
}

func (s *Service) CreateArticle(ctx context.Context, a domarticle.Article) (uint64, error) {
	id, err := s.repo.CreateArticle(ctx, &a)
	if err == nil {
		s.invalidateArticleCache(ctx)
	}
	return id, err
}

func (s *Service) UpdateArticle(ctx context.Context, id uint64, a domarticle.Article) error {
	a.ID = id
	err := s.repo.UpdateArticle(ctx, &a)
	if err == nil {
		s.invalidateArticleCache(ctx)
		go s.invalidateArticleDetailCache(ctx, id)
	}
	return err
}

func (s *Service) UpdateArticleStatus(ctx context.Context, id uint64, isPublished, isTop *bool) error {
	err := s.repo.UpdateArticleStatus(ctx, id, isPublished, isTop)
	if err == nil {
		s.invalidateArticleCache(ctx)
		go s.invalidateArticleDetailCache(ctx, id)
	}
	return err
}

func (s *Service) BatchUpdateArticleStatus(ctx context.Context, ids []uint64, isPublished, isTop *bool) error {
	err := s.repo.BatchUpdateArticleStatus(ctx, ids, isPublished, isTop)
	if err == nil {
		s.invalidateArticleCache(ctx)
	}
	return err
}

func (s *Service) DeleteArticle(ctx context.Context, id uint64) error {
	err := s.repo.DeleteArticle(ctx, id)
	if err == nil {
		s.invalidateArticleCache(ctx)
		go s.invalidateArticleDetailCache(ctx, id)
	}
	return err
}

func (s *Service) BatchDeleteArticles(ctx context.Context, ids []uint64) error {
	err := s.repo.BatchDeleteArticle(ctx, ids)
	if err == nil {
		s.invalidateArticleCache(ctx)
	}
	return err
}

func (s *Service) IncrementShareCount(ctx context.Context, id uint64) error {
	return s.repo.IncreaseShareCount(ctx, id)
}

func (s *Service) LikeArticle(ctx context.Context, id uint64) (int, error) {
	return s.repo.LikeArticle(ctx, id)
}

func (s *Service) RandomArticle(ctx context.Context, excludeID uint64) (*domarticle.Article, error) {
	return s.repo.RandomArticle(ctx, excludeID)
}

func (s *Service) GetArticleDetail(ctx context.Context, id uint64, increasePV bool) (*dto.ArticleDetailResponse, error) {
	if increasePV {
		_ = s.repo.IncreaseReadAndPV(ctx, id)
	}

	detailKey := fmt.Sprintf(cache.KeyArticleDetail, id)
	var result dto.ArticleDetailResponse
	err := s.cache.GetOrSet(ctx, detailKey, &result, func() (interface{}, error) {
		a, err := s.repo.FindByIDArticle(ctx, id)
		if err != nil {
			return nil, err
		}
		prev, _ := s.repo.FindPrev(ctx, id)
		next, _ := s.repo.FindNext(ctx, id)
		return &dto.ArticleDetailResponse{Article: a, PrevArticle: prev, NextArticle: next}, nil
	}, cache.LongTTLMin)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetArticleBySlug 根据 slug 获取文章详情（SEO 友好 URL）
func (s *Service) GetArticleBySlug(ctx context.Context, slug string) (*domarticle.Article, error) {
	// slug 查询不走缓存（slug 可能变更）
	article, err := s.repo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	// 增加 PV
	if article != nil {
		_ = s.repo.IncreaseReadAndPV(ctx, article.ID)
	}
	return article, nil
}

func (s *Service) Archives(ctx context.Context) (*dto.ListResponse[domarticle.ArchiveItem], error) {
	var result dto.ListResponse[domarticle.ArchiveItem]
	err := s.cache.GetOrSet(ctx, cache.KeyArchiveAll, &result, func() (interface{}, error) {
		items, err := s.repo.ArchiveList(ctx)
		if err != nil {
			return nil, err
		}
		return &dto.ListResponse[domarticle.ArchiveItem]{List: items}, nil
	}, cache.LongTTLMin)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) RecommendArticlesForMP(ctx context.Context, articleID uint64, size int) (*dto.PageResponse[domarticle.Article], error) {
	strategyName := "rule"
	if s.settingRepo != nil {
		st, err := s.settingRepo.Get(ctx)
		if err == nil && st != nil && st.UIConfig != nil {
			if name, ok := st.UIConfig["recommendStrategy"].(string); ok && name != "" {
				strategyName = name
			}
		}
	}

	var articles []domarticle.Article
	var err error

	if s.recRegistry != nil {
		strategy := s.recRegistry.Get(strategyName)
		if strategy != nil {
			articles, err = strategy.Recommend(ctx, recommendation.Context{
				CurrentArticleID: articleID,
				Limit:            size,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	if len(articles) == 0 {
		q := domarticle.ListQuery{Page: 1, Size: size, IsPublished: func() *bool { b := true; return &b }()}
		articles, _, err = s.repo.ListArticles(ctx, q)
		if err != nil {
			return nil, err
		}
	}

	return &dto.PageResponse[domarticle.Article]{
		List:  articles,
		Total: int64(len(articles)),
		Page:  1,
		Size:  len(articles),
	}, nil
}

func (s *Service) RecommendRelatedArticles(ctx context.Context, articleID uint64, size int) ([]domarticle.Article, error) {
	key := fmt.Sprintf(cache.KeyArticleRelated, articleID)
	var result []domarticle.Article
	err := s.cache.GetOrSet(ctx, key, &result, func() (interface{}, error) {
		return s.computeRelatedArticles(ctx, articleID, size)
	}, cache.DefaultTTLMin)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) computeRelatedArticles(ctx context.Context, articleID uint64, size int) ([]domarticle.Article, error) {
	detail, err := s.repo.FindByIDArticle(ctx, articleID)
	if err != nil {
		return nil, err
	}

	isPublished := true
	q := domarticle.ListQuery{
		Page:        1,
		Size:        size + 20,
		Keyword:     "",
		CategoryID:  0,
		TagID:       0,
		IsPublished: &isPublished,
	}
	list, _, err := s.repo.ListArticles(ctx, q)
	if err != nil {
		return nil, err
	}

	articleTags := make(map[uint64]bool)
	for _, tag := range detail.Tags {
		articleTags[tag.ID] = true
	}
	articleCategoryID := detail.CategoryID

	type scoredArticle struct {
		article domarticle.Article
		score   int
	}
	var scoredList []scoredArticle

	for _, article := range list {
		if article.ID == articleID {
			continue
		}
		score := 0
		if article.CategoryID == articleCategoryID {
			score += 3
		}
		for _, tag := range article.Tags {
			if articleTags[tag.ID] {
				score += 2
			}
		}
		if score > 0 || len(scoredList) < size {
			scoredList = append(scoredList, scoredArticle{article: article, score: score})
		}
	}

	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].score != scoredList[j].score {
			return scoredList[i].score > scoredList[j].score
		}
		return scoredList[i].article.ReadNum > scoredList[j].article.ReadNum
	})

	if len(scoredList) > size {
		scoredList = scoredList[:size]
	}

	result := make([]domarticle.Article, 0, len(scoredList))
	for _, sa := range scoredList {
		result = append(result, sa.article)
	}

	return result, nil
}

func (s *Service) ListArticlesByIDs(ctx context.Context, ids []uint64) ([]domarticle.Article, error) {
	return s.repo.ListArticlesByIDs(ctx, ids)
}

func (s *Service) ListAllPublishedArticles(ctx context.Context) ([]domarticle.Article, error) {
	isPublished := true
	q := domarticle.ListQuery{
		Page:        1,
		Size:        9999,
		Keyword:     "",
		CategoryID:  0,
		TagID:       0,
		IsPublished: &isPublished,
	}
	list, _, err := s.repo.ListArticles(ctx, q)
	return list, err
}

func (s *Service) GetRealHotSearches(ctx context.Context, limit int) ([]string, error) {
	var result []string
	err := s.cache.GetOrSet(ctx, cache.KeyHotSearches, &result, func() (interface{}, error) {
		return s.computeHotSearches(ctx, limit)
	}, cache.DefaultTTLMin)
	if err != nil {
		return nil, err
	}
	if len(result) > limit {
		return result[:limit], nil
	}
	return result, nil
}

func (s *Service) computeHotSearches(ctx context.Context, limit int) ([]string, error) {
	hotSearches, err := s.repo.GetHotSearchKeywords(ctx, limit)
	if err == nil && len(hotSearches) > 0 {
		return hotSearches, nil
	}

	hotTags, err := s.repo.GetHotTagsFromArticles(ctx, limit)
	if err == nil && len(hotTags) > 0 {
		return hotTags, nil
	}

	isPublished := true
	q := domarticle.ListQuery{
		Page:        1,
		Size:        limit,
		Keyword:     "",
		CategoryID:  0,
		TagID:       0,
		IsPublished: &isPublished,
	}
	list, _, err := s.repo.ListArticles(ctx, q)
	if err != nil {
		return nil, err
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ReadNum > list[j].ReadNum
	})

	result := make([]string, 0, limit)
	for i := 0; i < limit && i < len(list); i++ {
		if list[i].Title != "" {
			result = append(result, list[i].Title)
		}
	}

	return result, nil
}

func (s *Service) RecordSearchHistory(ctx context.Context, keyword string) error {
	return s.repo.RecordSearchHistory(ctx, keyword)
}

func (s *Service) ListAllPublishedArticlesForSitemap(ctx context.Context) ([]domarticle.Article, error) {
	var result []domarticle.Article
	err := s.cache.GetOrSet(ctx, cache.KeySitemapArticles, &result, func() (interface{}, error) {
		isPublished := true
		q := domarticle.ListQuery{
			Page:        1,
			Size:        9999,
			Keyword:     "",
			CategoryID:  0,
			TagID:       0,
			IsPublished: &isPublished,
		}
		list, _, err := s.repo.ListArticles(ctx, q)
		if err != nil {
			return nil, err
		}
		return list, nil
	}, cache.LongTTLMin)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// invalidateArticleCache 清除文章列表、归档、仪表盘等与文章相关的缓存。
func (s *Service) invalidateArticleCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	// 清除文章列表缓存
	if err := s.cache.DeletePattern(ctx, cache.KeyArticleListPattern); err != nil {
		logger.Warnf("清除文章列表缓存失败: %v", err)
	}
	// 清除归档缓存
	_ = s.cache.Delete(ctx, cache.KeyArchiveAll)
	// 清除仪表盘统计缓存
	_ = s.cache.Delete(ctx, cache.KeyDashboardStats)
	// 清除热门搜索缓存
	_ = s.cache.Delete(ctx, cache.KeyHotSearches)
	// 清除 sitemap 缓存
	_ = s.cache.Delete(ctx, cache.KeySitemapArticles)
}

// invalidateArticleDetailCache 清除单篇文章详情及前后篇缓存。
func (s *Service) invalidateArticleDetailCache(ctx context.Context, id uint64) {
	if s.cache == nil {
		return
	}
	detailKey := fmt.Sprintf(cache.KeyArticleDetail, id)
	prevKey := fmt.Sprintf(cache.KeyArticlePrev, id)
	nextKey := fmt.Sprintf(cache.KeyArticleNext, id)
	_ = s.cache.Delete(ctx, detailKey, prevKey, nextKey)
}
