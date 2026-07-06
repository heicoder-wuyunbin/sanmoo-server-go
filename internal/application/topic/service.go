package topic

import (
	"context"
	"fmt"
	"strings"

	domarticle "sanmoo-server-go/internal/domain/article"
	"sanmoo-server-go/internal/infrastructure/repository/mysqlrepo"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
)

type Service struct {
	repo *mysqlrepo.Repository
}

func NewService(repo *mysqlrepo.Repository) *Service {
	return &Service{repo: repo}
}

// ------ Admin CRUD ------

func (s *Service) ListAllTopicsWithCount(ctx context.Context) (*dto.ListResponse[dto.TopicItem], error) {
	topics, counts, err := s.repo.ListAllWithCountTopics(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]dto.TopicItem, 0, len(topics))
	for i, t := range topics {
		var c int64
		if i < len(counts) {
			c = counts[i]
		}
		list = append(list, dto.TopicItem{
			ID:           t.ID,
			Name:         t.Name,
			Description:  t.Description,
			CoverImage:   t.CoverImage,
			CreateTime:   t.CreateTime,
			ArticleCount: c,
		})
	}
	return &dto.ListResponse[dto.TopicItem]{List: list}, nil
}

func (s *Service) CreateTopic(ctx context.Context, name, description, coverImage string, articleIDs []uint64) (uint64, error) {
	id, err := s.repo.CreateTopic(ctx, name, description, coverImage)
	if err != nil {
		return 0, err
	}
	if len(articleIDs) > 0 {
		if err := s.repo.SetTopicArticles(ctx, id, articleIDs); err != nil {
			return id, err
		}
	}
	return id, nil
}

func (s *Service) UpdateTopic(ctx context.Context, id uint64, name, description, coverImage string, articleIDs []uint64) error {
	if err := s.repo.UpdateTopic(ctx, id, name, description, coverImage); err != nil {
		return err
	}
	return s.repo.SetTopicArticles(ctx, id, articleIDs)
}

func (s *Service) GetTopicArticleIDs(ctx context.Context, id uint64) ([]uint64, error) {
	return s.repo.GetTopicArticleIDs(ctx, id)
}

func (s *Service) ListPublishedArticleOptions(ctx context.Context) ([]dto.ArticleOption, error) {
	rows, err := s.repo.ListPublishedArticleOptions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ArticleOption, len(rows))
	for i, r := range rows {
		out[i] = dto.ArticleOption{ID: r.ID, Title: r.Title}
	}
	return out, nil
}

func (s *Service) DeleteTopic(ctx context.Context, id uint64) error {
	t, count, err := s.repo.FindByIDTopic(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperr.New(apperr.ErrConflict.Code, fmt.Sprintf("专题 \"%s\" 正在被 %d 篇文章使用，无法删除", t.Name, count))
	}
	return s.repo.DeleteTopic(ctx, id)
}

func (s *Service) BatchDeleteTopics(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	// 逐个检查并收集被使用的专题
	usedNames := make([]string, 0)
	for _, id := range ids {
		t, count, err := s.repo.FindByIDTopic(ctx, id)
		if err != nil {
			return err
		}
		if count > 0 {
			usedNames = append(usedNames, fmt.Sprintf("%s(%d篇)", t.Name, count))
		}
	}
	if len(usedNames) > 0 {
		return apperr.New(apperr.ErrConflict.Code, "以下专题正在被文章使用，无法删除："+strings.Join(usedNames, ", "))
	}
	return s.repo.BatchDeleteTopics(ctx, ids)
}

// ------ 小程序端查询 ------

func (s *Service) ListTopics(ctx context.Context, page, size int) (*dto.PageResponse[dto.TopicItem], error) {
	topics, counts, total, err := s.repo.ListTopics(ctx, page, size)
	if err != nil {
		return nil, err
	}
	list := make([]dto.TopicItem, 0, len(topics))
	for i, t := range topics {
		var c int64
		if i < len(counts) {
			c = counts[i]
		}
		list = append(list, dto.TopicItem{
			ID:           t.ID,
			Name:         t.Name,
			Description:  t.Description,
			CoverImage:   t.CoverImage,
			CreateTime:   t.CreateTime,
			ArticleCount: c,
		})
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	return &dto.PageResponse[dto.TopicItem]{List: list, Total: total, Page: page, Size: size}, nil
}

func (s *Service) GetTopicDetail(ctx context.Context, id uint64) (*dto.TopicDetailResponse, error) {
	t, count, err := s.repo.GetTopicByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.TopicDetailResponse{
		Topic: dto.TopicItem{
			ID:           t.ID,
			Name:         t.Name,
			Description:  t.Description,
			CoverImage:   t.CoverImage,
			CreateTime:   t.CreateTime,
			ArticleCount: count,
		},
	}, nil
}

func (s *Service) ListTopicArticles(ctx context.Context, topicID uint64, page, size int) (*dto.PageResponse[domarticle.Article], error) {
	one := true
	list, total, err := s.repo.ListArticlesByTopic(ctx, topicID, page, size, &one)
	if err != nil {
		return nil, err
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	return &dto.PageResponse[domarticle.Article]{List: list, Total: total, Page: page, Size: size}, nil
}
