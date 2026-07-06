package link

type LinkRepository interface {
	List(page, size int, keyword string) ([]*Link, int64, error)
	ListActive() ([]*Link, error)
	GetByID(id uint64) (*Link, error)
	Create(link *Link) error
	Update(link *Link) error
	Delete(id uint64) error
	BatchDelete(ids []uint64) error
}