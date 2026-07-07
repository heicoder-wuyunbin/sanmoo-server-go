package role

import (
	"context"
	"fmt"
	"strings"

	domperm "sanmoo-server-go/internal/domain/permission"
	domrole "sanmoo-server-go/internal/domain/role"
	"sanmoo-server-go/internal/infrastructure/cache"
	"sanmoo-server-go/internal/interfaces/http/dto"
	apperr "sanmoo-server-go/internal/shared/errors"
)

type Service struct {
	roleRepo domrole.Repository
	permRepo domperm.Repository
	cache    *cache.BusinessCache
}

func NewService(roleRepo domrole.Repository, permRepo domperm.Repository, cache *cache.BusinessCache) *Service {
	return &Service{roleRepo: roleRepo, permRepo: permRepo, cache: cache}
}

const (
	cacheKeyRoleList     = "blog:role:all"
	cacheKeyRolePermsFmt = "blog:role:%d:perms"
	cacheKeyUserRolesFmt = "blog:user:%d:roles"
	cacheTTLMin          = 30
)

func (s *Service) ListRoles(ctx context.Context, page, size int, keyword string) (*dto.PageResponse[domrole.Role], error) {
	q := domrole.ListQuery{Page: page, Size: size, Keyword: keyword}
	list, total, err := s.roleRepo.ListRoles(ctx, q)
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
	return &dto.PageResponse[domrole.Role]{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *Service) ListAllRoles(ctx context.Context) (*dto.ListResponse[domrole.Role], error) {
	var cached dto.ListResponse[domrole.Role]
	if s.cache != nil {
		if hit, _ := s.cache.Get(ctx, cacheKeyRoleList, &cached); hit {
			return &cached, nil
		}
	}
	list, err := s.roleRepo.ListAllRoles(ctx)
	if err != nil {
		return nil, err
	}
	result := &dto.ListResponse[domrole.Role]{List: list}
	if s.cache != nil {
		_ = s.cache.SetWithTTL(ctx, cacheKeyRoleList, result, cacheTTLMin)
	}
	return result, nil
}

func (s *Service) GetRole(ctx context.Context, id uint64) (*domrole.RoleWithPermissions, error) {
	role, err := s.roleRepo.FindByIDRole(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	permKeys, err := s.roleRepo.GetRolePermKeys(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domrole.RoleWithPermissions{Role: *role, PermKeys: permKeys}, nil
}

func (s *Service) CreateRole(ctx context.Context, r *domrole.Role) (uint64, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}
	existing, _ := s.roleRepo.FindByName(ctx, r.Name)
	if existing != nil {
		return 0, apperr.New(apperr.ErrConflict.Code, "角色名已存在")
	}
	id, err := s.roleRepo.CreateRole(ctx, r)
	if err == nil {
		s.invalidateRoleCache(ctx)
	}
	return id, err
}

func (s *Service) UpdateRole(ctx context.Context, r *domrole.Role) error {
	if err := r.Validate(); err != nil {
		return err
	}
	_, err := s.roleRepo.FindByIDRole(ctx, r.ID)
	if err != nil {
		return apperr.ErrNotFound
	}
	err = s.roleRepo.UpdateRole(ctx, r)
	if err == nil {
		s.invalidateRoleCache(ctx)
	}
	return err
}

func (s *Service) DeleteRole(ctx context.Context, id uint64) error {
	role, err := s.roleRepo.FindByIDRole(ctx, id)
	if err != nil {
		return apperr.ErrNotFound
	}
	if s.roleRepo.IsAdminRole(role.Name) {
		return apperr.New(apperr.ErrForbidden.Code, "不能删除 admin 角色")
	}
	err = s.roleRepo.DeleteRole(ctx, id)
	if err == nil {
		s.invalidateRoleCache(ctx)
		s.invalidateRolePermCache(ctx, id)
	}
	return err
}

func (s *Service) AssignPermissions(ctx context.Context, roleID uint64, permKeys []string) error {
	_, err := s.roleRepo.FindByIDRole(ctx, roleID)
	if err != nil {
		return apperr.ErrNotFound
	}
	for _, key := range permKeys {
		p, err := s.permRepo.FindByPermKey(ctx, key)
		if err != nil || p == nil {
			return apperr.New(apperr.ErrInvalidParam.Code, fmt.Sprintf("权限 %s 不存在", key))
		}
	}
	err = s.roleRepo.AssignRolePermissions(ctx, roleID, permKeys)
	if err == nil {
		s.invalidateRolePermCache(ctx, roleID)
	}
	return err
}

func (s *Service) GetRolePermKeys(ctx context.Context, roleID uint64) ([]string, error) {
	cacheKey := fmt.Sprintf(cacheKeyRolePermsFmt, roleID)
	if s.cache != nil {
		var cached []string
		if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
			return cached, nil
		}
	}
	keys, err := s.roleRepo.GetRolePermKeys(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetWithTTL(ctx, cacheKey, keys, cacheTTLMin)
	}
	return keys, nil
}

func (s *Service) GetUserRoles(ctx context.Context, userID uint64) ([]domrole.Role, error) {
	cacheKey := fmt.Sprintf(cacheKeyUserRolesFmt, userID)
	if s.cache != nil {
		var cached []domrole.Role
		if hit, _ := s.cache.Get(ctx, cacheKey, &cached); hit {
			return cached, nil
		}
	}
	roles, err := s.roleRepo.GetRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetWithTTL(ctx, cacheKey, roles, cacheTTLMin)
	}
	return roles, nil
}

func (s *Service) GetUserPermKeys(ctx context.Context, userID uint64) (map[string]bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	permSet := make(map[string]bool)
	roleIDs := make([]uint64, 0, len(roles))
	for _, r := range roles {
		if s.roleRepo.IsAdminRole(r.Name) {
			allPerms, err := s.permRepo.ListAllPermissions(ctx)
			if err != nil {
				return nil, err
			}
			for _, p := range allPerms {
				permSet[p.PermKey] = true
			}
			return permSet, nil
		}
		roleIDs = append(roleIDs, r.ID)
	}
	if len(roleIDs) == 0 {
		return permSet, nil
	}
	keys, err := s.permRepo.GetPermKeysByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		permSet[k] = true
	}
	return permSet, nil
}

func (s *Service) HasPermission(ctx context.Context, userID uint64, permKey string) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if s.roleRepo.IsAdminRole(r.Name) {
			return true, nil
		}
	}
	permSet, err := s.GetUserPermKeys(ctx, userID)
	if err != nil {
		return false, err
	}
	return permSet[permKey], nil
}

func (s *Service) AssignUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	err := s.roleRepo.AssignUserRoles(ctx, userID, roleIDs)
	if err == nil {
		s.invalidateUserRoleCache(ctx, userID)
	}
	return err
}

// GetUserMenus 获取用户的菜单列表（用于前端动态菜单渲染）
func (s *Service) GetUserMenus(ctx context.Context, userID uint64) ([]domperm.UserMenuItem, error) {
	permSet, err := s.GetUserPermKeys(ctx, userID)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(permSet))
	for k := range permSet {
		keys = append(keys, k)
	}
	return s.permRepo.GetUserMenus(ctx, keys)
}

func (s *Service) IsAdminUser(ctx context.Context, userID uint64) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if s.roleRepo.IsAdminRole(r.Name) {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) invalidateRoleCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	_ = s.cache.Delete(ctx, cacheKeyRoleList)
}

func (s *Service) invalidateRolePermCache(ctx context.Context, roleID uint64) {
	if s.cache == nil {
		return
	}
	cacheKey := fmt.Sprintf(cacheKeyRolePermsFmt, roleID)
	_ = s.cache.Delete(ctx, cacheKey)
	_ = s.cache.DeletePattern(ctx, "blog:user:*:roles")
}

func (s *Service) invalidateUserRoleCache(ctx context.Context, userID uint64) {
	if s.cache == nil {
		return
	}
	cacheKey := fmt.Sprintf(cacheKeyUserRolesFmt, userID)
	_ = s.cache.Delete(ctx, cacheKey)
}

func IsAdminRoleName(name string) bool {
	return strings.EqualFold(name, "admin") || strings.EqualFold(name, "ROLE_ADMIN")
}
