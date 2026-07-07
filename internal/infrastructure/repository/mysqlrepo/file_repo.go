package mysqlrepo

import (
	"bytes"
	"context"
	"crypto/md5"
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
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	domfile "sanmoo-server-go/internal/domain/file"
	apperr "sanmoo-server-go/internal/shared/errors"
)

func (r *Repository) List(ctx context.Context, q domfile.ListQuery) ([]domfile.FileItem, int64, error) {
	strategy := r.getUploadStrategy()
	if strategy == "QINIU" {
		return r.listQiniuFiles(ctx, q)
	}

	dir := r.getUploadLocalDir()
	if dir == "" {
		dir = "uploads"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, 0, err
	}

	items := make([]domfile.FileItem, 0)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		name := d.Name()
		if q.Keyword != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(q.Keyword)) {
			return nil
		}
		info, _ := d.Info()
		items = append(items, domfile.FileItem{
			ID:         fileID(relativePath),
			Name:       relativePath,
			Size:       info.Size(),
			URL:        joinURL(r.getUploadURLPrefix(), relativePath),
			CreateTime: info.ModTime().Format("2006-01-02 15:04:05"),
		})
		return nil
	})
	if err != nil {
		return nil, 0, err
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

func (r *Repository) listQiniuFiles(ctx context.Context, q domfile.ListQuery) ([]domfile.FileItem, int64, error) {
	return []domfile.FileItem{}, 0, nil
}

func (r *Repository) Upload(ctx context.Context, fileHeader *multipart.FileHeader) (domfile.FileItem, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return domfile.FileItem{}, err
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		return domfile.FileItem{}, err
	}
	return r.UploadBytes(ctx, fileHeader.Filename, data)
}

func (r *Repository) UploadBytes(ctx context.Context, filename string, data []byte) (domfile.FileItem, error) {
	strategy := r.getUploadStrategy()
	if strategy == "QINIU" {
		return r.uploadToQiniu(filename, data)
	}

	dir := r.getUploadLocalDir()
	if dir == "" {
		dir = "uploads"
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}

	h := md5.Sum(data)
	md5Str := hex.EncodeToString(h[:])
	safeName := md5Str + ext

	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())

	fullDir := filepath.Join(dir, datePath)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return domfile.FileItem{}, err
	}

	compressedData := compressImageIfNeeded(safeName, data)

	dstPath := filepath.Join(fullDir, safeName)
	if err := os.WriteFile(dstPath, compressedData, 0644); err != nil {
		return domfile.FileItem{}, err
	}

	relativePath := filepath.Join(datePath, safeName)
	return domfile.FileItem{
		ID:         fileID(relativePath),
		Name:       relativePath,
		Size:       int64(len(compressedData)),
		URL:        joinURL(r.getUploadURLPrefix(), relativePath),
		CreateTime: now.Format("2006-01-02 15:04:05"),
	}, nil
}

func (r *Repository) uploadToQiniu(filename string, data []byte) (domfile.FileItem, error) {
	accessKey, secretKey, bucket, _ := r.getQiniuConfig()
	if accessKey == "" || secretKey == "" || bucket == "" {
		return domfile.FileItem{}, apperr.New(apperr.ErrInternal.Code, "七牛云存储配置未完成")
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}

	h := md5.Sum(data)
	md5Str := hex.EncodeToString(h[:])
	safeName := md5Str + ext

	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	relativePath := filepath.Join(datePath, safeName)

	compressedData := compressImageIfNeeded(safeName, data)

	mac := qbox.NewMac(accessKey, secretKey)
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{}
	cfg.Zone = &storage.ZoneHuanan
	cfg.UseHTTPS = false

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}

	err := formUploader.Put(context.Background(), &ret, upToken, relativePath, bytes.NewReader(compressedData), int64(len(compressedData)), &putExtra)
	if err != nil {
		return domfile.FileItem{}, err
	}

	return domfile.FileItem{
		ID:         fileID(relativePath),
		Name:       relativePath,
		Size:       int64(len(compressedData)),
		URL:        r.getQiniuFileURL(relativePath),
		CreateTime: now.Format("2006-01-02 15:04:05"),
	}, nil
}

func (r *Repository) getQiniuFileURL(filename string) string {
	accessKey, secretKey, _, domain := r.getQiniuConfig()
	if domain == "" {
		domain = r.getUploadURLPrefix()
	}
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "http://" + domain
	}

	mac := qbox.NewMac(accessKey, secretKey)
	deadline := time.Now().Add(3600 * time.Second).Unix()
	privateURL := storage.MakePrivateURL(mac, domain, filename, deadline)
	return privateURL
}

func (r *Repository) DeleteByID(ctx context.Context, id string) error {
	strategy := r.getUploadStrategy()
	if strategy == "QINIU" {
		return r.deleteFromQiniu(ctx, id)
	}

	dir := r.getUploadLocalDir()
	if dir == "" {
		dir = "uploads"
	}

	var foundPath string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		if fileID(relativePath) == id {
			foundPath = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return err
	}
	if foundPath == "" {
		return apperr.ErrNotFound
	}
	return os.Remove(foundPath)
}

func (r *Repository) deleteFromQiniu(ctx context.Context, id string) error {
	accessKey, secretKey, bucket, _ := r.getQiniuConfig()
	if accessKey == "" || secretKey == "" || bucket == "" {
		return apperr.New(apperr.ErrInternal.Code, "七牛云存储配置未完成")
	}

	dir := r.getUploadLocalDir()
	if dir == "" {
		dir = "uploads"
	}

	var relativePath string
	var fullPath string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rp, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		if fileID(rp) == id {
			relativePath = rp
			fullPath = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return err
	}
	if relativePath == "" {
		return apperr.ErrNotFound
	}

	mac := qbox.NewMac(accessKey, secretKey)
	cfg := storage.Config{}
	cfg.Zone = &storage.ZoneHuanan
	cfg.UseHTTPS = false
	bucketManager := storage.NewBucketManager(mac, &cfg)

	if err := bucketManager.Delete(bucket, relativePath); err != nil {
		return err
	}

	return os.Remove(fullPath)
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
