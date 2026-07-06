package mysqlrepo

import (
	"sanmoo-server-go/internal/domain/link"

	"gorm.io/gorm"
)

type linkRepo struct {
	db *gorm.DB
}

func NewLinkRepo(db *gorm.DB) link.LinkRepository {
	return &linkRepo{db: db}
}

func toLink(m *tLink) *link.Link {
	return &link.Link{
		ID:          m.ID,
		Name:        m.Name,
		Url:         m.Url,
		Description: m.Description,
		Icon:        m.Icon,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreateTime:  m.CreateTime,
		UpdateTime:  m.UpdateTime,
	}
}

func toLinkList(ms []*tLink) []*link.Link {
	res := make([]*link.Link, 0, len(ms))
	for _, m := range ms {
		res = append(res, toLink(m))
	}
	return res
}

func (r *linkRepo) List(page, size int, keyword string) ([]*link.Link, int64, error) {
	var models []*tLink
	var total int64

	query := r.db.Model(&tLink{}).Order("sort_order ASC, create_time DESC")

	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset((page - 1) * size).Limit(size).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	return toLinkList(models), total, nil
}

func (r *linkRepo) ListActive() ([]*link.Link, error) {
	var models []*tLink
	err := r.db.Model(&tLink{}).
		Where("is_active = ?", true).
		Order("sort_order ASC, create_time DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toLinkList(models), nil
}

func (r *linkRepo) GetByID(id uint64) (*link.Link, error) {
	var m tLink
	err := r.db.First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return toLink(&m), nil
}

func (r *linkRepo) Create(l *link.Link) error {
	m := tLink{
		Name:        l.Name,
		Url:         l.Url,
		Description: l.Description,
		Icon:        l.Icon,
		SortOrder:   l.SortOrder,
		IsActive:    l.IsActive,
	}
	err := r.db.Create(&m).Error
	if err != nil {
		return err
	}
	l.ID = m.ID
	l.CreateTime = m.CreateTime
	l.UpdateTime = m.UpdateTime
	return nil
}

func (r *linkRepo) Update(l *link.Link) error {
	m := tLink{
		ID:          l.ID,
		Name:        l.Name,
		Url:         l.Url,
		Description: l.Description,
		Icon:        l.Icon,
		SortOrder:   l.SortOrder,
		IsActive:    l.IsActive,
	}
	err := r.db.Save(&m).Error
	if err != nil {
		return err
	}
	l.UpdateTime = m.UpdateTime
	return nil
}

func (r *linkRepo) Delete(id uint64) error {
	return r.db.Delete(&tLink{}, id).Error
}

func (r *linkRepo) BatchDelete(ids []uint64) error {
	return r.db.Delete(&tLink{}, ids).Error
}
