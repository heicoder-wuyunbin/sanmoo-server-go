package file

import "mime/multipart"
import "context"

type ListQuery struct {
	Page    int
	Size    int
	Keyword string
}

type Repository interface {
	List(ctx context.Context, q ListQuery) ([]FileItem, int64, error)
	Upload(ctx context.Context, fileHeader *multipart.FileHeader) (FileItem, error)
	UploadBytes(ctx context.Context, filename string, data []byte) (FileItem, error)
	DeleteByID(ctx context.Context, id string) error
}
