package file

import (
	"context"
	"mime/multipart"
	domfile "sanmoo-server-go/internal/domain/file"
	"sanmoo-server-go/internal/interfaces/http/dto"
	"sanmoo-server-go/internal/shared/pagination"
)

type Service struct {
	repo domfile.Repository
}

func NewService(repo domfile.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListFiles(ctx context.Context, page, size int, keyword string) (*pagination.PageData, error) {
	list, total, err := s.repo.List(ctx, domfile.ListQuery{Page: page, Size: size, Keyword: keyword})
	if err != nil {
		return nil, err
	}
	return pagination.NewPageData(list, total, page, size), nil
}

func (s *Service) UploadFile(ctx context.Context, file *multipart.FileHeader) (*dto.FileUploadResponse, error) {
	item, err := s.repo.Upload(ctx, file)
	if err != nil {
		return nil, err
	}
	return &dto.FileUploadResponse{URL: item.URL, Size: item.Size}, nil
}

func (s *Service) UploadFileBytes(ctx context.Context, filename string, data []byte) (*dto.FileUploadResponse, error) {
	item, err := s.repo.UploadBytes(ctx, filename, data)
	if err != nil {
		return nil, err
	}
	return &dto.FileUploadResponse{URL: item.URL, Size: item.Size}, nil
}

func (s *Service) DeleteFile(ctx context.Context, id string) error {
	return s.repo.DeleteByID(ctx, id)
}

func (s *Service) GetProxyURL(ctx context.Context, filePath string) (string, error) {
	return s.repo.GetProxyURL(ctx, filePath)
}
