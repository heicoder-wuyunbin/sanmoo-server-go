package mysqlrepo

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	domfile "sanmoo-server-go/internal/domain/file"
	apperr "sanmoo-server-go/internal/shared/errors"
)

func (r *Repository) List(ctx context.Context, q domfile.ListQuery) ([]domfile.FileItem, int64, error) {
	dir := r.getUploadLocalDir()
	if dir == "" {
		dir = "uploads"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, 0, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0, err
	}
	items := make([]domfile.FileItem, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if q.Keyword != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(q.Keyword)) {
			continue
		}
		info, _ := e.Info()
		items = append(items, domfile.FileItem{ID: fileID(name), Name: name, Size: info.Size(), URL: joinURL(r.getUploadURLPrefix(), name), CreateTime: info.ModTime().Format("2006-01-02 15:04:05")})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreateTime > items[j].CreateTime })
	total := int64(len(items))
	start := (q.Page - 1) * q.Size
	if start >= len(items) {
		return []domfile.FileItem{}, total, nil
	}
	end := start + q.Size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func (r *Repository) Upload(ctx context.Context, fileHeader *multipart.FileHeader) (domfile.FileItem, error) {
	dir := r.getUploadLocalDir()
	if dir == "" {
		dir = "uploads"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return domfile.FileItem{}, err
	}
	src, err := fileHeader.Open()
	if err != nil {
		return domfile.FileItem{}, err
	}
	defer src.Close()
	filename := fmt.Sprintf("%d_%s", time.Now().UnixMilli(), filepath.Base(fileHeader.Filename))
	data, err := io.ReadAll(src)
	if err != nil {
		return domfile.FileItem{}, err
	}
	return r.UploadBytes(ctx, filename, data)
}

func (r *Repository) UploadBytes(ctx context.Context, filename string, data []byte) (domfile.FileItem, error) {
	dir := r.getUploadLocalDir()
	if dir == "" {
		dir = "uploads"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return domfile.FileItem{}, err
	}
	safeName := fmt.Sprintf("%d_%s", time.Now().UnixMilli(), filepath.Base(filename))
	if safeName == "" || safeName == "." {
		safeName = fmt.Sprintf("%d_upload.bin", time.Now().UnixMilli())
	}

	compressedData := compressImageIfNeeded(safeName, data)

	dstPath := filepath.Join(dir, safeName)
	if err := os.WriteFile(dstPath, compressedData, 0644); err != nil {
		return domfile.FileItem{}, err
	}
	return domfile.FileItem{
		ID:         fileID(safeName),
		Name:       safeName,
		Size:       int64(len(compressedData)),
		URL:        joinURL(r.getUploadURLPrefix(), safeName),
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func compressImageIfNeeded(filename string, data []byte) []byte {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return data
	}

	if len(data) < 1024*100 {
		return data
	}

	var img image.Image
	var err error
	if ext == ".png" {
		img, err = png.Decode(bytes.NewReader(data))
	} else {
		img, err = jpeg.Decode(bytes.NewReader(data))
	}
	if err != nil {
		return data
	}

	maxWidth := 1920
	maxHeight := 1080
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	if width > maxWidth || height > maxHeight {
		img = imaging.Resize(img, maxWidth, maxHeight, imaging.Lanczos)
	}

	var buf bytes.Buffer
	quality := 80
	if len(data) > 1024*1024*2 {
		quality = 70
	}

	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})

	if err != nil || buf.Len() >= len(data) {
		return data
	}

	return buf.Bytes()
}

func (r *Repository) DeleteByID(ctx context.Context, id string) error {
	dir := r.getUploadLocalDir()
	if dir == "" {
		dir = "uploads"
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if fileID(e.Name()) == id {
			return os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	return apperr.ErrNotFound
}

func fileID(name string) string {
	h := sha1.Sum([]byte(name))
	return hex.EncodeToString(h[:8])
}
func joinURL(prefix, name string) string {
	p := prefix
	if p == "" {
		p = "/uploads/"
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p + name
}
