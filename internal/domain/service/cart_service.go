package service

import (
	"context"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
)

type CartService struct {
	repo        *repository.CartRepo
	productRepo *repository.ProductRepo
}

func NewCartService(cartRepo *repository.CartRepo, productRepo *repository.ProductRepo) *CartService {
	return &CartService{repo: cartRepo, productRepo: productRepo}
}

type CartItemVO struct {
	ID          uint64 `json:"id"`
	ProductID   uint64 `json:"product_id"`
	SKUID       uint64 `json:"sku_id"`
	ProductName string `json:"product_name"`
	Spec        string `json:"spec"`
	CoverImage  string `json:"cover_image"`
	Price       int64  `json:"price"`
	Stock       int64  `json:"stock"`
	Quantity    int    `json:"quantity"`
	Checked     bool   `json:"checked"`
}

func (s *CartService) List(ctx context.Context, tenantID, userID uint64) ([]CartItemVO, error) {
	items, err := s.repo.List(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	vos := make([]CartItemVO, 0, len(items))
	for _, item := range items {
		sku, err := s.productRepo.GetSKU(ctx, tenantID, item.SkuID)
		if err != nil {
			continue
		}
		p, _ := s.productRepo.GetByID(ctx, tenantID, item.ProductID)
		name, cover := "", ""
		if p != nil {
			name = p.Name
			cover = p.CoverImage
		}
		vos = append(vos, CartItemVO{
			ID: item.ID, ProductID: item.ProductID, SKUID: item.SkuID,
			ProductName: name, Spec: sku.Spec, CoverImage: cover, Price: sku.Price, Stock: sku.Stock,
			Quantity: item.Quantity, Checked: item.Checked,
		})
	}
	return vos, nil
}

func (s *CartService) Add(ctx context.Context, tenantID, userID, productID, skuID uint64, qty int) error {
	if qty < 1 {
		qty = 1
	}
	return s.repo.Upsert(ctx, &model.CartItem{
		TenantScoped: model.TenantScoped{TenantID: tenantID},
		UserID: userID, ProductID: productID, SkuID: skuID, Quantity: qty, Checked: true,
	})
}

func (s *CartService) Remove(ctx context.Context, tenantID, userID, itemID uint64) error {
	return s.repo.Delete(ctx, tenantID, userID, itemID)
}

func (s *CartService) Update(ctx context.Context, tenantID, userID, itemID uint64, quantity *int, checked *bool) error {
	return s.repo.Update(ctx, tenantID, userID, itemID, quantity, checked)
}

func (s *CartService) CheckAll(ctx context.Context, tenantID, userID uint64, checked bool) error {
	return s.repo.CheckAll(ctx, tenantID, userID, checked)
}
