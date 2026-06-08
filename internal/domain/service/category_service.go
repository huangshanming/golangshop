package service

import (
	"context"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepo
}

func NewCategoryService(repo *repository.CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context, tenantID uint64) ([]model.ProductCategory, error) {
	return s.repo.List(ctx, tenantID)
}
