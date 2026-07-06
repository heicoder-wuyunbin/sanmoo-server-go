package link

import (
	"sanmoo-server-go/internal/domain/link"
	apperr "sanmoo-server-go/internal/shared/errors"
)

type LinkService struct {
	repo link.LinkRepository
}

func NewLinkService(repo link.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

func (s *LinkService) List(page, size int, keyword string) ([]*link.Link, int64, error) {
	return s.repo.List(page, size, keyword)
}

func (s *LinkService) ListActive() ([]*link.Link, error) {
	return s.repo.ListActive()
}

func (s *LinkService) GetByID(id uint64) (*link.Link, error) {
	return s.repo.GetByID(id)
}

func (s *LinkService) Create(name, url, description, icon string, sortOrder int) (*link.Link, error) {
	l := &link.Link{
		Name:        name,
		Url:         url,
		Description: description,
		Icon:        icon,
		SortOrder:   sortOrder,
		IsActive:    true,
	}

	if err := l.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(l); err != nil {
		return nil, apperr.New(apperr.ErrInternal.Code, "创建友情链接失败")
	}

	return l, nil
}

func (s *LinkService) Update(id uint64, name, url, description, icon string, sortOrder int, isActive bool) (*link.Link, error) {
	l, err := s.repo.GetByID(id)
	if err != nil {
		return nil, apperr.New(apperr.ErrNotFound.Code, "友情链接不存在")
	}

	l.Name = name
	l.Url = url
	l.Description = description
	l.Icon = icon
	l.SortOrder = sortOrder
	l.IsActive = isActive

	if err := l.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(l); err != nil {
		return nil, apperr.New(apperr.ErrInternal.Code, "更新友情链接失败")
	}

	return l, nil
}

func (s *LinkService) Delete(id uint64) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return apperr.New(apperr.ErrNotFound.Code, "友情链接不存在")
	}

	return s.repo.Delete(id)
}

func (s *LinkService) BatchDelete(ids []uint64) error {
	return s.repo.BatchDelete(ids)
}