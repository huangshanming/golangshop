package repository

import (
	"context"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type CartRepo struct {
	db *gorm.DB
}

func NewCartRepo(db *gorm.DB) *CartRepo { return &CartRepo{db: db} }

func (r *CartRepo) List(ctx context.Context, tenantID, userID uint64) ([]model.CartItem, error) {
	var items []model.CartItem
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Find(&items).Error
	return items, err
}

func (r *CartRepo) Upsert(ctx context.Context, item *model.CartItem) error {
	var existing model.CartItem
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND sku_id = ?",
		item.TenantID, item.UserID, item.SkuID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(item).Error
	}
	if err != nil {
		return err
	}
	existing.Quantity = item.Quantity
	existing.Checked = item.Checked
	return r.db.WithContext(ctx).Save(&existing).Error
}

func (r *CartRepo) Delete(ctx context.Context, tenantID, userID, itemID uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, itemID).
		Delete(&model.CartItem{}).Error
}

func (r *CartRepo) GetByID(ctx context.Context, tenantID, userID, itemID uint64) (*model.CartItem, error) {
	var item model.CartItem
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, itemID).First(&item).Error
	return &item, err
}

func (r *CartRepo) Update(ctx context.Context, tenantID, userID, itemID uint64, quantity *int, checked *bool) error {
	updates := map[string]interface{}{}
	if quantity != nil {
		updates["quantity"] = *quantity
	}
	if checked != nil {
		updates["checked"] = *checked
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.CartItem{}).
		Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, itemID).
		Updates(updates).Error
}

func (r *CartRepo) CheckAll(ctx context.Context, tenantID, userID uint64, checked bool) error {
	return r.db.WithContext(ctx).Model(&model.CartItem{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Update("checked", checked).Error
}

func (r *CartRepo) DeleteByIDs(ctx context.Context, tenantID, userID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND id IN ?", tenantID, userID, ids).
		Delete(&model.CartItem{}).Error
}

func (r *CartRepo) DeleteByIDsWithTx(tx *gorm.DB, tenantID, userID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.Where("tenant_id = ? AND user_id = ? AND id IN ?", tenantID, userID, ids).
		Delete(&model.CartItem{}).Error
}
