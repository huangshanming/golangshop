package repository

import (
	"context"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type ProductRepo struct {
	db *gorm.DB
}

func NewProductRepo(db *gorm.DB) *ProductRepo { return &ProductRepo{db: db} }

type ProductListFilter struct {
	CategoryID uint64
	Keyword    string
	OnlyOnSale bool
}

func (r *ProductRepo) applyFilter(q *gorm.DB, tenantID uint64, f ProductListFilter) *gorm.DB {
	q = q.Where("tenant_id = ?", tenantID)
	if f.OnlyOnSale {
		q = q.Where("status = 1")
	}
	if f.CategoryID > 0 {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("name LIKE ? OR subtitle LIKE ?", like, like)
	}
	return q
}

func (r *ProductRepo) List(ctx context.Context, tenantID uint64, page, pageSize int, f ProductListFilter) ([]model.Product, int64, error) {
	var list []model.Product
	var total int64
	q := r.applyFilter(r.db.WithContext(ctx).Model(&model.Product{}), tenantID, f)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

func (r *ProductRepo) GetByID(ctx context.Context, tenantID, id uint64) (*model.Product, error) {
	var p model.Product
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&p).Error
	return &p, err
}

func (r *ProductRepo) Create(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *ProductRepo) Update(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Omit("CreatedAt").Save(p).Error
}

func (r *ProductRepo) UpdateStatus(ctx context.Context, tenantID, id uint64, status int8) error {
	return r.db.WithContext(ctx).Model(&model.Product{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("status", status).Error
}

func (r *ProductRepo) ListSKUByProduct(ctx context.Context, tenantID, productID uint64) ([]model.ProductSKU, error) {
	var skus []model.ProductSKU
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND product_id = ? AND status = 1", tenantID, productID).
		Find(&skus).Error
	return skus, err
}

func (r *ProductRepo) ListSKUByProductAll(ctx context.Context, tenantID, productID uint64) ([]model.ProductSKU, error) {
	var skus []model.ProductSKU
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Find(&skus).Error
	return skus, err
}

func (r *ProductRepo) ListSKUByProducts(ctx context.Context, tenantID uint64, productIDs []uint64) (map[uint64][]model.ProductSKU, error) {
	result := make(map[uint64][]model.ProductSKU)
	if len(productIDs) == 0 {
		return result, nil
	}
	var skus []model.ProductSKU
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND product_id IN ? AND status = 1", tenantID, productIDs).
		Find(&skus).Error
	if err != nil {
		return nil, err
	}
	for _, sku := range skus {
		result[sku.ProductID] = append(result[sku.ProductID], sku)
	}
	return result, nil
}

func (r *ProductRepo) ListSKUByProductsAll(ctx context.Context, tenantID uint64, productIDs []uint64) (map[uint64][]model.ProductSKU, error) {
	result := make(map[uint64][]model.ProductSKU)
	if len(productIDs) == 0 {
		return result, nil
	}
	var skus []model.ProductSKU
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND product_id IN ?", tenantID, productIDs).
		Find(&skus).Error
	if err != nil {
		return nil, err
	}
	for _, sku := range skus {
		result[sku.ProductID] = append(result[sku.ProductID], sku)
	}
	return result, nil
}

func (r *ProductRepo) GetSKU(ctx context.Context, tenantID, skuID uint64) (*model.ProductSKU, error) {
	var sku model.ProductSKU
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, skuID).First(&sku).Error
	return &sku, err
}

func (r *ProductRepo) CreateSKU(ctx context.Context, sku *model.ProductSKU) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

func (r *ProductRepo) ReplaceSKUs(ctx context.Context, tenantID, productID uint64, skus []model.ProductSKU) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND product_id = ?", tenantID, productID).
			Delete(&model.ProductSKU{}).Error; err != nil {
			return err
		}
		for i := range skus {
			skus[i].ProductID = productID
			skus[i].TenantID = tenantID
			if err := tx.Create(&skus[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProductRepo) DeductStock(ctx context.Context, tenantID, skuID uint64, qty int, version int64) (int64, error) {
	return r.deductStock(r.db.WithContext(ctx), tenantID, skuID, qty, version)
}

func (r *ProductRepo) DeductStockWithTx(tx *gorm.DB, tenantID, skuID uint64, qty int, version int64) (int64, error) {
	return r.deductStock(tx, tenantID, skuID, qty, version)
}

func (r *ProductRepo) deductStock(db *gorm.DB, tenantID, skuID uint64, qty int, version int64) (int64, error) {
	res := db.Model(&model.ProductSKU{}).
		Where("tenant_id = ? AND id = ? AND version = ? AND stock >= ?", tenantID, skuID, version, qty).
		Updates(map[string]interface{}{
			"stock":   gorm.Expr("stock - ?", qty),
			"version": version + 1,
		})
	return res.RowsAffected, res.Error
}

func (r *ProductRepo) RestoreStock(ctx context.Context, tenantID, skuID uint64, qty int) error {
	return r.restoreStock(r.db.WithContext(ctx), tenantID, skuID, qty)
}

func (r *ProductRepo) RestoreStockWithTx(tx *gorm.DB, tenantID, skuID uint64, qty int) error {
	return r.restoreStock(tx, tenantID, skuID, qty)
}

func (r *ProductRepo) restoreStock(db *gorm.DB, tenantID, skuID uint64, qty int) error {
	return db.Model(&model.ProductSKU{}).
		Where("tenant_id = ? AND id = ?", tenantID, skuID).
		UpdateColumn("stock", gorm.Expr("stock + ?", qty)).Error
}
