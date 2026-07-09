package permission

import (
	"context"

	domperm "sanmoo-server-go/internal/domain/permission"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/infrastructure/repository/mysqlrepo"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
)

type Service struct {
	repo  domperm.Repository
	cache *cache.BusinessCache
}

func NewService(repo domperm.Repository, cache *cache.BusinessCache) *Service {
	return &Service{repo: repo, cache: cache}
}

const (
	cacheKeyAllPerms     = "blog:perm:all"
	cacheKeyPermTree     = "blog:perm:tree"
	cacheKeyRolePermsFmt = "blog:role:%d:perms"
	cacheTTLMin          = 30
)

func (s *Service) ListPermissions(ctx context.Context, page, size int, keyword, module, ptype string) (*dto.PageResponse[domperm.Permission], error) {
	q := domperm.ListQuery{Page: page, Size: size, Keyword: keyword, Module: module, Type: ptype}
	list, total, err := s.repo.ListPermissions(ctx, q)
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
	return &dto.PageResponse[domperm.Permission]{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *Service) ListAllPermissions(ctx context.Context) (*dto.ListResponse[domperm.Permission], error) {
	var cached dto.ListResponse[domperm.Permission]
	if s.cache != nil {
		if hit, _ := s.cache.Get(ctx, cacheKeyAllPerms, &cached); hit {
			return &cached, nil
		}
	}
	list, err := s.repo.ListAllPermissions(ctx)
	if err != nil {
		return nil, err
	}
	result := &dto.ListResponse[domperm.Permission]{List: list}
	if s.cache != nil {
		_ = s.cache.SetWithTTL(ctx, cacheKeyAllPerms, result, cacheTTLMin)
	}
	return result, nil
}

func (s *Service) PermissionTree(ctx context.Context) ([]domperm.PermissionTree, error) {
	all, err := s.ListAllPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return mysqlrepo.BuildPermissionTree(all.List), nil
}

func (s *Service) GetPermission(ctx context.Context, id uint64) (*domperm.Permission, error) {
	return s.repo.FindByIDPermission(ctx, id)
}

func (s *Service) CreatePermission(ctx context.Context, p *domperm.Permission) (uint64, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}
	existing, _ := s.repo.FindByPermKey(ctx, p.PermKey)
	if existing != nil {
		return 0, apperr.New(apperr.ErrConflict.Code, "权限标识已存在")
	}
	id, err := s.repo.CreatePermission(ctx, p)
	if err == nil {
		s.invalidatePermCache(ctx)
	}
	return id, err
}

// FROZEN (L3): 权限管理已冻结，不再允许更新权限
func (s *Service) UpdatePermission(ctx context.Context, p *domperm.Permission) error {
	if err := p.Validate(); err != nil {
		return err
	}
	_, err := s.repo.FindByIDPermission(ctx, p.ID)
	if err != nil {
		return apperr.ErrNotFound
	}
	err = s.repo.UpdatePermission(ctx, p)
	if err == nil {
		s.invalidatePermCache(ctx)
	}
	return err
}

// FROZEN (L3): 权限管理已冻结，不再允许删除权限
func (s *Service) DeletePermission(ctx context.Context, id uint64) error {
	_, err := s.repo.FindByIDPermission(ctx, id)
	if err != nil {
		return apperr.ErrNotFound
	}
	err = s.repo.DeletePermission(ctx, id)
	if err == nil {
		s.invalidatePermCache(ctx)
	}
	return err
}

func (s *Service) invalidatePermCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	_ = s.cache.Delete(ctx, cacheKeyAllPerms, cacheKeyPermTree)
	_ = s.cache.DeletePattern(ctx, "blog:role:*:perms")
}
