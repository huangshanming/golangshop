package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type CouponRepo struct {
	db *gorm.DB
}

func NewCouponRepo(db *gorm.DB) *CouponRepo { return &CouponRepo{db: db} }

func (r *CouponRepo) GetByID(ctx context.Context, tenantID, id uint64) (*model.Coupon, error) {
	var c model.Coupon
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&c).Error
	return &c, err
}

func (r *CouponRepo) ListActive(ctx context.Context, tenantID uint64) ([]model.Coupon, error) {
	now := time.Now().Unix()
	var list []model.Coupon
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND status = 1 AND valid_from <= ? AND valid_to >= ?", tenantID, now, now).
		Find(&list).Error
	return list, err
}

func (r *CouponRepo) Create(ctx context.Context, c *model.Coupon) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CouponRepo) ListAll(ctx context.Context, tenantID uint64) ([]model.Coupon, error) {
	var list []model.Coupon
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("id DESC").Find(&list).Error
	return list, err
}

func (r *CouponRepo) ListPage(ctx context.Context, tenantID uint64, page, pageSize int, keyword string) ([]model.Coupon, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := r.db.WithContext(ctx).Model(&model.Coupon{}).Where("tenant_id = ?", tenantID)
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Coupon
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *CouponRepo) Update(ctx context.Context, c *model.Coupon) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *CouponRepo) UpdateStatus(ctx context.Context, tenantID, id uint64, status int8) error {
	return r.db.WithContext(ctx).Model(&model.Coupon{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("status", status).Error
}

func (r *CouponRepo) GetUserCoupon(ctx context.Context, tenantID, userID, userCouponID uint64) (*model.UserCoupon, error) {
	var uc model.UserCoupon
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, userCouponID).First(&uc).Error
	return &uc, err
}

func (r *CouponRepo) ListUserCoupons(ctx context.Context, tenantID, userID uint64, status string) ([]model.UserCoupon, error) {
	var list []model.UserCoupon
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *CouponRepo) CountUserCoupon(ctx context.Context, tenantID, userID, couponID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("tenant_id = ? AND user_id = ? AND coupon_id = ?", tenantID, userID, couponID).Count(&count).Error
	return count, err
}

func (r *CouponRepo) CreateUserCoupon(ctx context.Context, uc *model.UserCoupon) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Coupon{}).
			Where("id = ? AND (total_count = 0 OR received_count < total_count)", uc.CouponID).
			UpdateColumn("received_count", gorm.Expr("received_count + 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(uc).Error
	})
}

func (r *CouponRepo) MarkUserCouponUsed(ctx context.Context, tenantID, userCouponID uint64, orderNo string, usedAt int64) error {
	return r.markUserCouponUsed(r.db.WithContext(ctx), tenantID, userCouponID, orderNo, usedAt)
}

func (r *CouponRepo) MarkUserCouponUsedWithTx(tx *gorm.DB, tenantID, userCouponID uint64, orderNo string, usedAt int64) error {
	return r.markUserCouponUsed(tx, tenantID, userCouponID, orderNo, usedAt)
}

func (r *CouponRepo) markUserCouponUsed(db *gorm.DB, tenantID, userCouponID uint64, orderNo string, usedAt int64) error {
	return db.Model(&model.UserCoupon{}).
		Where("tenant_id = ? AND id = ? AND status = ?", tenantID, userCouponID, model.UserCouponUnused).
		Updates(map[string]interface{}{"status": model.UserCouponUsed, "used_order_no": orderNo, "used_at": usedAt}).Error
}

func (r *CouponRepo) RestoreUserCoupon(ctx context.Context, tenantID, userCouponID uint64) error {
	return r.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("tenant_id = ? AND id = ?", tenantID, userCouponID).
		Updates(map[string]interface{}{"status": model.UserCouponUnused, "used_order_no": "", "used_at": nil}).Error
}

func (r *CouponRepo) CreateOrderPromotion(ctx context.Context, p *model.OrderPromotion) error {
	return r.db.WithContext(ctx).Create(p).Error
}
