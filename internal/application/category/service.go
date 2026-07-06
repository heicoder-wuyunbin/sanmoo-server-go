package category

import (
	"context"
	"fmt"
	"strings"

	domcategory "sanmoo-server-go/internal/domain/category"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
)

type Service struct {
	repo  domcategory.Repository
	cache *cache.BusinessCache
}

func NewService(repo domcategory.Repository, cache *cache.BusinessCache) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) ListCategories(ctx context.Context) (*dto.ListResponse[domcategory.Category], error) {
	// 尝试从缓存读取
	if s.cache != nil {
		var cached dto.ListResponse[domcategory.Category]
		if hit, _ := s.cache.Get(ctx, cache.KeyCategoryAll, &cached); hit {
			return &cached, nil
		}
	}

	list, err := s.repo.ListAllWithCountCategory(ctx)
	if err != nil {
		return nil, err
	}
	result := &dto.ListResponse[domcategory.Category]{List: list}

	// 写入缓存
	if s.cache != nil {
		_ = s.cache.SetWithTTL(ctx, cache.KeyCategoryAll, result, cache.LongTTLMin)
	}

	return result, nil
}

func (s *Service) CreateCategory(ctx context.Context, name string) (uint64, error) {
	c := domcategory.Category{Name: name}
	id, err := s.repo.CreateCategory(ctx, &c)
	if err == nil {
		s.invalidateCategoryCache(ctx)
	}
	return id, err
}

func (s *Service) UpdateCategory(ctx context.Context, id uint64, name string) error {
	c := domcategory.Category{ID: id, Name: name}
	err := s.repo.UpdateCategory(ctx, &c)
	if err == nil {
		s.invalidateCategoryCache(ctx)
	}
	return err
}

func (s *Service) DeleteCategory(ctx context.Context, id uint64) error {
	c, err := s.repo.FindByIDCategory(ctx, id)
	if err != nil {
		return err
	}
	if c.ArticleCount > 0 {
		return apperr.New(apperr.ErrConflict.Code, fmt.Sprintf("分类 \"%s\" 正在被 %d 篇文章使用，无法删除", c.Name, c.ArticleCount))
	}
	err = s.repo.DeleteCategory(ctx, id)
	if err == nil {
		s.invalidateCategoryCache(ctx)
	}
	return err
}

func (s *Service) BatchDeleteCategories(ctx context.Context, ids []uint64) error {
	categories, err := s.repo.ListCategoriesByIDs(ctx, ids)
	if err != nil {
		return err
	}
	usedNames := make([]string, 0)
	for _, c := range categories {
		if c.ArticleCount > 0 {
			usedNames = append(usedNames, fmt.Sprintf("%s(%d篇)", c.Name, c.ArticleCount))
		}
	}
	if len(usedNames) > 0 {
		return apperr.New(apperr.ErrConflict.Code, "以下分类正在被文章使用，无法删除："+strings.Join(usedNames, ", "))
	}
	err = s.repo.BatchDeleteCategory(ctx, ids)
	if err == nil {
		s.invalidateCategoryCache(ctx)
	}
	return err
}

func (s *Service) invalidateCategoryCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	_ = s.cache.Delete(ctx, cache.KeyCategoryAll)
}
