package tag

import (
	"context"
	"fmt"
	"strings"

	domtag "sanmoo-server-go/internal/domain/tag"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
)

type Service struct {
	repo  domtag.Repository
	cache *cache.BusinessCache
}

func NewService(repo domtag.Repository, cache *cache.BusinessCache) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) ListTags(ctx context.Context, page, size int, keyword string) (*dto.PageResponse[domtag.Tag], error) {
	q := domtag.ListQuery{Page: page, Size: size, Keyword: keyword}
	list, total, err := s.repo.ListTags(ctx, q)
	if err != nil {
		return nil, err
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = len(list)
		if size <= 0 {
			size = 10
		}
	}
	return &dto.PageResponse[domtag.Tag]{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *Service) ListAllTags(ctx context.Context) (*dto.ListResponse[domtag.Tag], error) {
	// 尝试从缓存读取
	if s.cache != nil {
		var cached dto.ListResponse[domtag.Tag]
		if hit, _ := s.cache.Get(ctx, cache.KeyTagAll, &cached); hit {
			return &cached, nil
		}
	}

	list, err := s.repo.ListAllWithCountTag(ctx)
	if err != nil {
		return nil, err
	}
	result := &dto.ListResponse[domtag.Tag]{List: list}

	// 写入缓存
	if s.cache != nil {
		_ = s.cache.SetWithTTL(ctx, cache.KeyTagAll, result, cache.LongTTLMin)
	}

	return result, nil
}

func (s *Service) CreateTag(ctx context.Context, name string) (uint64, error) {
	t := domtag.Tag{Name: name}
	id, err := s.repo.CreateTag(ctx, &t)
	if err == nil {
		s.invalidateTagCache(ctx)
	}
	return id, err
}

func (s *Service) UpdateTag(ctx context.Context, id uint64, name string) error {
	t := domtag.Tag{ID: id, Name: name}
	err := s.repo.UpdateTag(ctx, &t)
	if err == nil {
		s.invalidateTagCache(ctx)
	}
	return err
}

func (s *Service) DeleteTag(ctx context.Context, id uint64) error {
	t, err := s.repo.FindByIDTag(ctx, id)
	if err != nil {
		return err
	}
	if t.ArticleCount > 0 {
		return apperr.New(apperr.ErrConflict.Code, fmt.Sprintf("标签 \"%s\" 正在被 %d 篇文章使用，无法删除", t.Name, t.ArticleCount))
	}
	err = s.repo.DeleteTag(ctx, id)
	if err == nil {
		s.invalidateTagCache(ctx)
	}
	return err
}

func (s *Service) BatchDeleteTags(ctx context.Context, ids []uint64) error {
	tags, err := s.repo.ListTagsByIDs(ctx, ids)
	if err != nil {
		return err
	}
	usedNames := make([]string, 0)
	for _, t := range tags {
		if t.ArticleCount > 0 {
			usedNames = append(usedNames, fmt.Sprintf("%s(%d篇)", t.Name, t.ArticleCount))
		}
	}
	if len(usedNames) > 0 {
		return apperr.New(apperr.ErrConflict.Code, "以下标签正在被文章使用，无法删除："+strings.Join(usedNames, ", "))
	}
	err = s.repo.BatchDeleteTag(ctx, ids)
	if err == nil {
		s.invalidateTagCache(ctx)
	}
	return err
}

func (s *Service) invalidateTagCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	_ = s.cache.Delete(ctx, cache.KeyTagAll)
}
