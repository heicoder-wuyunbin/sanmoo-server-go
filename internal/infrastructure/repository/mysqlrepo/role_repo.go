package mysqlrepo

import (
	"context"

	domrole "sanmoo-server-go/internal/domain/role"

	"gorm.io/gorm"
)

func (r *Repository) ListRoles(ctx context.Context, q domrole.ListQuery) ([]domrole.Role, int64, error) {
	db := r.db.WithContext(ctx).Table("t_role").Where("status = 1")
	if q.Keyword != "" {
		db = db.Where("(name like ? OR description like ?)", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var models []tRole
	query := r.db.WithContext(ctx).Table("t_role").Where("status = 1")
	if q.Keyword != "" {
		query = query.Where("(name like ? OR description like ?)", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	query = query.Order("sort_order asc, id asc")
	if q.Page > 0 && q.Size > 0 {
		query = query.Offset((q.Page - 1) * q.Size).Limit(q.Size)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, 0, err
	}
	res := make([]domrole.Role, 0, len(models))
	for _, m := range models {
		res = append(res, tRoleToDomain(&m))
	}
	return res, total, nil
}

func (r *Repository) ListAllRoles(ctx context.Context) ([]domrole.Role, error) {
	var models []tRole
	err := r.db.WithContext(ctx).Table("t_role").
		Where("status = 1").
		Order("sort_order asc, id asc").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	res := make([]domrole.Role, 0, len(models))
	for _, m := range models {
		res = append(res, tRoleToDomain(&m))
	}
	return res, nil
}

func (r *Repository) FindByIDRole(ctx context.Context, id uint64) (*domrole.Role, error) {
	var m tRole
	if err := r.db.WithContext(ctx).Where("status = 1").First(&m, id).Error; err != nil {
		return nil, err
	}
	r2 := tRoleToDomain(&m)
	return &r2, nil
}

func (r *Repository) FindByName(ctx context.Context, name string) (*domrole.Role, error) {
	var m tRole
	if err := r.db.WithContext(ctx).Where("name = ? AND status = 1", name).First(&m).Error; err != nil {
		return nil, err
	}
	r2 := tRoleToDomain(&m)
	return &r2, nil
}

func (r *Repository) CreateRole(ctx context.Context, role *domrole.Role) (uint64, error) {
	m := tRole{
		Name:        role.Name,
		Description: role.Description,
		Status:      role.Status,
		SortOrder:   role.SortOrder,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (r *Repository) UpdateRole(ctx context.Context, role *domrole.Role) error {
	updates := map[string]any{
		"name":        role.Name,
		"description": role.Description,
		"sort_order":  role.SortOrder,
		"status":      role.Status,
	}
	return r.db.WithContext(ctx).Model(&tRole{}).Where("id = ?", role.ID).Updates(updates).Error
}

func (r *Repository) DeleteRole(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&tRole{}).Where("id = ?", id).Update("status", 0).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&tRolePermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&tUserRole{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) GetRolesByUserID(ctx context.Context, userID uint64) ([]domrole.Role, error) {
	var models []tRole
	err := r.db.WithContext(ctx).Table("t_role").
		Joins("JOIN t_user_role ON t_user_role.role_id = t_role.id").
		Where("t_user_role.user_id = ? AND t_role.status = 1", userID).
		Order("t_role.sort_order asc, t_role.id asc").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	res := make([]domrole.Role, 0, len(models))
	for _, m := range models {
		res = append(res, tRoleToDomain(&m))
	}
	return res, nil
}

func (r *Repository) GetRoleIDsByUserID(ctx context.Context, userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Table("t_user_role").
		Where("user_id = ?", userID).
		Pluck("role_id", &ids).Error
	return ids, err
}

func (r *Repository) AssignRolePermissions(ctx context.Context, roleID uint64, permKeys []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&tRolePermission{}).Error; err != nil {
			return err
		}
		if len(permKeys) == 0 {
			return nil
		}
		records := make([]tRolePermission, 0, len(permKeys))
		for _, key := range permKeys {
			records = append(records, tRolePermission{RoleID: roleID, PermKey: key})
		}
		return tx.Create(&records).Error
	})
}

func (r *Repository) GetRolePermKeys(ctx context.Context, roleID uint64) ([]string, error) {
	var keys []string
	err := r.db.WithContext(ctx).Table("t_role_permission").
		Where("role_id = ?", roleID).
		Pluck("perm_key", &keys).Error
	return keys, err
}

func (r *Repository) AssignUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&tUserRole{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		records := make([]tUserRole, 0, len(roleIDs))
		for _, rid := range roleIDs {
			records = append(records, tUserRole{UserID: userID, RoleID: rid})
		}
		return tx.Create(&records).Error
	})
}

func (r *Repository) IsAdminRole(roleName string) bool {
	return IsAdminRole(roleName)
}

func tRoleToDomain(m *tRole) domrole.Role {
	return domrole.Role{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Status:      m.Status,
		SortOrder:   m.SortOrder,
		CreateTime:  m.CreateTime,
		UpdateTime:  m.UpdateTime,
	}
}
