package service

import (
	"context"
)

type CategoryService struct {
	platform *PlatformService
}

func NewCategoryService(platform *PlatformService) *CategoryService {
	return &CategoryService{platform: platform}
}

func (s *CategoryService) ListTree(ctx context.Context) ([]PlatformCategoryTreeItem, error) {
	return s.platform.ListCategoryTree(ctx, true)
}
