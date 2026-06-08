package repository

import (
	"context"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type AfterSaleRepo struct {
	db *gorm.DB
}

func NewAfterSaleRepo(db *gorm.DB) *AfterSaleRepo { return &AfterSaleRepo{db: db} }

func (r *AfterSaleRepo) Create(ctx context.Context, a *model.AfterSaleOrder) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *AfterSaleRepo) GetByNo(ctx context.Context, tenantID uint64, afterSaleNo string) (*model.AfterSaleOrder, error) {
	var a model.AfterSaleOrder
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND after_sale_no = ?", tenantID, afterSaleNo).First(&a).Error
	return &a, err
}

func (r *AfterSaleRepo) ListByUser(ctx context.Context, tenantID, userID uint64, page, pageSize int) ([]model.AfterSaleOrder, int64, error) {
	var list []model.AfterSaleOrder
	var total int64
	q := r.db.WithContext(ctx).Model(&model.AfterSaleOrder{}).Where("tenant_id = ? AND user_id = ?", tenantID, userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

func (r *AfterSaleRepo) HasActiveByOrder(ctx context.Context, tenantID, orderID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AfterSaleOrder{}).
		Where("tenant_id = ? AND order_id = ? AND status IN ?", tenantID, orderID,
			[]string{model.AfterSaleStatusPending, model.AfterSaleStatusApproved, model.AfterSaleStatusProcessing}).
		Count(&count).Error
	return count > 0, err
}

func (r *AfterSaleRepo) UpdateStatus(ctx context.Context, tenantID uint64, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&model.AfterSaleOrder{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).Update("status", status).Error
}
