package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"goshop/internal/infra/storage"
)

const maxUploadSize = 5 << 20 // 5MB

var allowedExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

type UploadService struct {
	storage *storage.MinIOClient
}

func NewUploadService(s *storage.MinIOClient) *UploadService {
	return &UploadService{storage: s}
}

type UploadResult struct {
	URL  string `json:"url"`
	Path string `json:"path"`
}

func (s *UploadService) UploadImage(ctx context.Context, tenantID uint64, filename string, reader io.Reader, size int64) (*UploadResult, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("文件存储未配置")
	}
	if size <= 0 || size > maxUploadSize {
		return nil, fmt.Errorf("文件大小需在 1B ~ 5MB 之间")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	contentType, ok := allowedExt[ext]
	if !ok {
		return nil, fmt.Errorf("仅支持 jpg/png/webp 图片")
	}
	if ct := mime.TypeByExtension(ext); ct != "" {
		contentType = ct
	}
	objectKey := fmt.Sprintf("products/%d/%s%s", tenantID, uuid.New().String(), ext)
	url, err := s.storage.PutObject(ctx, objectKey, reader, size, contentType)
	if err != nil {
		return nil, err
	}
	return &UploadResult{URL: url, Path: objectKey}, nil
}
