package mysqlrepo

import (
	"context"
	"strings"

	domperm "sanmoo-server-go/internal/domain/permission"
)

func (r *Repository) ListPermissions(ctx context.Context, q domperm.ListQuery) ([]domperm.Permission, int64, error) {
	db := r.db.WithContext(ctx).Table("t_permission").Where("status = 1")
	if q.Keyword != "" {
		db = db.Where("(name like ? OR perm_key like ?)", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	if q.Module != "" {
		db = db.Where("module = ?", q.Module)
	}
	if q.Type != "" {
		db = db.Where("type = ?", q.Type)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var models []tPermission
	query := r.db.WithContext(ctx).Table("t_permission").Where("status = 1")
	if q.Keyword != "" {
		query = query.Where("(name like ? OR perm_key like ?)", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	if q.Module != "" {
		query = query.Where("module = ?", q.Module)
	}
	if q.Type != "" {
		query = query.Where("type = ?", q.Type)
	}
	query = query.Order("sort_order asc, id asc")
	if q.Page > 0 && q.Size > 0 {
		query = query.Offset((q.Page - 1) * q.Size).Limit(q.Size)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, 0, err
	}
	res := make([]domperm.Permission, 0, len(models))
	for _, m := range models {
		res = append(res, tPermissionToDomain(&m))
	}
	return res, total, nil
}

func (r *Repository) ListAllPermissions(ctx context.Context) ([]domperm.Permission, error) {
	var models []tPermission
	err := r.db.WithContext(ctx).Table("t_permission").
		Where("status = 1").
		Order("sort_order asc, id asc").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	res := make([]domperm.Permission, 0, len(models))
	for _, m := range models {
		res = append(res, tPermissionToDomain(&m))
	}
	return res, nil
}

func (r *Repository) FindByIDPermission(ctx context.Context, id uint64) (*domperm.Permission, error) {
	var m tPermission
	if err := r.db.WithContext(ctx).Where("status = 1").First(&m, id).Error; err != nil {
		return nil, err
	}
	p := tPermissionToDomain(&m)
	return &p, nil
}

func (r *Repository) FindByPermKey(ctx context.Context, permKey string) (*domperm.Permission, error) {
	var m tPermission
	if err := r.db.WithContext(ctx).Where("perm_key = ? AND status = 1", permKey).First(&m).Error; err != nil {
		return nil, err
	}
	p := tPermissionToDomain(&m)
	return &p, nil
}

func (r *Repository) CreatePermission(ctx context.Context, p *domperm.Permission) (uint64, error) {
	m := tPermission{
		PermKey:     p.PermKey,
		Name:        p.Name,
		Module:      p.Module,
		Type:        p.Type,
		Description: p.Description,
		SortOrder:   p.SortOrder,
		Status:      p.Status,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (r *Repository) UpdatePermission(ctx context.Context, p *domperm.Permission) error {
	updates := map[string]any{
		"name":        p.Name,
		"module":      p.Module,
		"type":        p.Type,
		"description": p.Description,
		"sort_order":  p.SortOrder,
		"status":      p.Status,
	}
	return r.db.WithContext(ctx).Model(&tPermission{}).Where("id = ?", p.ID).Updates(updates).Error
}

func (r *Repository) DeletePermission(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&tPermission{}).Where("id = ?", id).Update("status", 0).Error
}

func (r *Repository) GetPermKeysByRoleIDs(ctx context.Context, roleIDs []uint64) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	var keys []string
	err := r.db.WithContext(ctx).Table("t_role_permission").
		Where("role_id IN ?", roleIDs).
		Pluck("perm_key", &keys).Error
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func tPermissionToDomain(m *tPermission) domperm.Permission {
	return domperm.Permission{
		ID:          m.ID,
		PermKey:     m.PermKey,
		Name:        m.Name,
		Module:      m.Module,
		Type:        m.Type,
		Description: m.Description,
		SortOrder:   m.SortOrder,
		Status:      m.Status,
		CreateTime:  m.CreateTime,
		UpdateTime:  m.UpdateTime,
	}
}

func BuildPermissionTree(perms []domperm.Permission) []domperm.PermissionTree {
	moduleMap := make(map[string]*domperm.PermissionTree)
	order := make([]string, 0)
	moduleNames := map[string]string{
		"dashboard":  "仪表盘",
		"article":    "文章管理",
		"category":   "分类管理",
		"tag":        "标签管理",
		"topic":      "专题管理",
		"link":       "友情链接",
		"file":       "文件管理",
		"user":       "用户管理",
		"setting":    "系统设置",
		"cache":      "缓存管理",
		"backup":     "备份管理",
		"mpuser":     "微信用户",
		"permission": "权限管理",
		"role":       "角色管理",
	}
	for _, p := range perms {
		tree, ok := moduleMap[p.Module]
		if !ok {
			name := p.Module
			if mn, ok := moduleNames[p.Module]; ok {
				name = mn
			}
			tree = &domperm.PermissionTree{
				Module:     p.Module,
				ModuleName: name,
				Children:   make([]domperm.PermissionTreeItem, 0),
			}
			moduleMap[p.Module] = tree
			order = append(order, p.Module)
		}
		tree.Children = append(tree.Children, domperm.PermissionTreeItem{
			ID:          p.ID,
			PermKey:     p.PermKey,
			Name:        p.Name,
			Type:        p.Type,
			Description: p.Description,
		})
	}
	result := make([]domperm.PermissionTree, 0, len(order))
	for _, mod := range order {
		result = append(result, *moduleMap[mod])
	}
	return result
}

func IsAdminRole(roleName string) bool {
	return strings.EqualFold(roleName, "admin") || strings.EqualFold(roleName, "ROLE_ADMIN")
}
