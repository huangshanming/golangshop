package service

import (
	"context"
	"errors"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
)

type ProductService struct {
	repo         *repository.ProductRepo
	categoryRepo *repository.CategoryRepo
}

func NewProductService(repo *repository.ProductRepo, categoryRepo *repository.CategoryRepo) *ProductService {
	return &ProductService{repo: repo, categoryRepo: categoryRepo}
}

type ProductListItem struct {
	model.Product
	MinPrice      int64  `json:"min_price"`
	MaxPrice      int64  `json:"max_price"`
	OriginalPrice int64  `json:"original_price"`
	TotalStock    int64  `json:"total_stock"`
	CategoryName  string `json:"category_name"`
}

type SKUInputItem struct {
	SpecValues    map[string]string
	Spec          string
	Price         int64
	OriginalPrice int64
	Stock         int64
	Image         string
}

func MarshalSpecsDimensions(dims []SpecDimension) string {
	return marshalSpecs(dims)
}

func BuildSKUsFromDimensions(dims []SpecDimension, items []SKUInputItem) ([]model.ProductSKU, error) {
	specsRaw := marshalSpecs(dims)
	return buildSKUsFromInput(specsRaw, items)
}

func (s *ProductService) buildListItems(ctx context.Context, tenantID uint64, products []model.Product, includeOffSaleSKUs bool) ([]ProductListItem, error) {
	if len(products) == 0 {
		return []ProductListItem{}, nil
	}
	ids := make([]uint64, 0, len(products))
	categoryIDs := make([]uint64, 0)
	for _, p := range products {
		ids = append(ids, p.ID)
		if p.CategoryID > 0 {
			categoryIDs = append(categoryIDs, p.CategoryID)
		}
	}
	var skuMap map[uint64][]model.ProductSKU
	var err error
	if includeOffSaleSKUs {
		skuMap, err = s.repo.ListSKUByProductsAll(ctx, tenantID, ids)
	} else {
		skuMap, err = s.repo.ListSKUByProducts(ctx, tenantID, ids)
	}
	if err != nil {
		return nil, err
	}
	categoryNames, err := s.categoryRepo.GetNamesByIDs(ctx, tenantID, categoryIDs)
	if err != nil {
		return nil, err
	}

	items := make([]ProductListItem, 0, len(products))
	for _, p := range products {
		item := ProductListItem{Product: p, CategoryName: categoryNames[p.CategoryID]}
		skus := skuMap[p.ID]
		for _, sku := range skus {
			if item.MinPrice == 0 || sku.Price < item.MinPrice {
				item.MinPrice = sku.Price
				item.OriginalPrice = sku.OriginalPrice
			}
			if sku.Price > item.MaxPrice {
				item.MaxPrice = sku.Price
			}
			item.TotalStock += sku.Stock
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *ProductService) ListApp(ctx context.Context, tenantID uint64, page, pageSize int, categoryID uint64, keyword string) ([]ProductListItem, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	list, total, err := s.repo.List(ctx, tenantID, page, pageSize, repository.ProductListFilter{
		CategoryID: categoryID,
		Keyword:    keyword,
		OnlyOnSale: true,
	})
	if err != nil {
		return nil, 0, err
	}
	items, err := s.buildListItems(ctx, tenantID, list, false)
	return items, total, err
}

func (s *ProductService) ListAdmin(ctx context.Context, tenantID uint64, page, pageSize int, keyword string) ([]ProductListItem, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	list, total, err := s.repo.List(ctx, tenantID, page, pageSize, repository.ProductListFilter{
		Keyword:    keyword,
		OnlyOnSale: false,
	})
	if err != nil {
		return nil, 0, err
	}
	items, err := s.buildListItems(ctx, tenantID, list, true)
	return items, total, err
}

func (s *ProductService) Detail(ctx context.Context, tenantID, productID uint64) (*model.Product, []model.ProductSKU, error) {
	p, err := s.repo.GetByID(ctx, tenantID, productID)
	if err != nil {
		return nil, nil, err
	}
	skus, err := s.repo.ListSKUByProduct(ctx, tenantID, productID)
	return p, skus, err
}

func (s *ProductService) AdminDetail(ctx context.Context, tenantID, productID uint64) (*model.Product, []model.ProductSKU, error) {
	p, err := s.repo.GetByID(ctx, tenantID, productID)
	if err != nil {
		return nil, nil, err
	}
	skus, err := s.repo.ListSKUByProductAll(ctx, tenantID, productID)
	return p, skus, err
}

func (s *ProductService) Create(ctx context.Context, p *model.Product, skus []model.ProductSKU) error {
	if len(skus) == 0 {
		return errors.New("至少需要一个 SKU")
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return err
	}
	for i := range skus {
		skus[i].ProductID = p.ID
		skus[i].TenantID = p.TenantID
		skus[i].Status = 1
		if err := s.repo.CreateSKU(ctx, &skus[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProductService) Update(ctx context.Context, tenantID, productID uint64, p *model.Product, skus []model.ProductSKU) error {
	existing, err := s.repo.GetByID(ctx, tenantID, productID)
	if err != nil {
		return err
	}
	if len(skus) == 0 {
		return errors.New("至少需要一个 SKU")
	}
	existing.CategoryID = p.CategoryID
	existing.Name = p.Name
	existing.Subtitle = p.Subtitle
	existing.Description = p.Description
	existing.CoverImage = p.CoverImage
	existing.Images = p.Images
	existing.Specs = p.Specs
	if err := s.repo.Update(ctx, existing); err != nil {
		return err
	}
	for i := range skus {
		skus[i].Status = 1
	}
	return s.repo.ReplaceSKUs(ctx, tenantID, productID, skus)
}

func (s *ProductService) UpdateStatus(ctx context.Context, tenantID, productID uint64, status int8) error {
	if _, err := s.repo.GetByID(ctx, tenantID, productID); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, tenantID, productID, status)
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
