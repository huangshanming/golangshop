package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"goshop/internal/domain/model"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo { return &OrderRepo{db: db} }

func (r *OrderRepo) DB() *gorm.DB { return r.db }

func (r *OrderRepo) CreateWithTx(tx *gorm.DB, order *model.Order) error {
	items := order.Items
	promotions := order.Promotions
	order.Items = nil
	order.Promotions = nil

	if err := tx.Omit(clause.Associations).Create(order).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].ID = 0
		items[i].OrderID = order.ID
		items[i].TenantID = order.TenantID
		if err := tx.Create(&items[i]).Error; err != nil {
			return err
		}
	}
	for i := range promotions {
		promotions[i].ID = 0
		promotions[i].OrderID = order.ID
		promotions[i].TenantID = order.TenantID
		if err := tx.Create(&promotions[i]).Error; err != nil {
			return err
		}
	}
	order.Items = items
	order.Promotions = promotions
	return nil
}

func (r *OrderRepo) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.CreateWithTx(tx, order)
	})
}

func (r *OrderRepo) FindByOrderNo(ctx context.Context, tenantID uint64, orderNo string) (*model.Order, error) {
	var o model.Order
	err := r.db.WithContext(ctx).Preload("Items").Preload("Promotions").
		Where("tenant_id = ? AND order_no = ?", tenantID, orderNo).First(&o).Error
	return &o, err
}

func (r *OrderRepo) FindByOrderNoForUser(ctx context.Context, tenantID, userID uint64, orderNo string) (*model.Order, error) {
	var o model.Order
	err := r.db.WithContext(ctx).Preload("Items").Preload("Promotions").
		Where("tenant_id = ? AND user_id = ? AND order_no = ?", tenantID, userID, orderNo).First(&o).Error
	return &o, err
}

func (r *OrderRepo) FindByClientOrderNo(ctx context.Context, tenantID uint64, clientOrderNo string) (*model.Order, error) {
	var o model.Order
	err := r.db.WithContext(ctx).Preload("Items").Preload("Promotions").
		Where("tenant_id = ? AND client_order_no = ?", tenantID, clientOrderNo).First(&o).Error
	return &o, err
}

func (r *OrderRepo) UpdatePaid(ctx context.Context, tenantID uint64, orderNo, payNo, channel string, payAmount int64, paidAt int64) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("tenant_id = ? AND order_no = ? AND status = ?", tenantID, orderNo, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":      model.OrderStatusPaid,
			"pay_no":      payNo,
			"pay_channel": channel,
			"pay_amount":  payAmount,
			"paid_at":     paidAt,
		}).Error
}

func (r *OrderRepo) UpdatePaidWithTx(tx *gorm.DB, tenantID uint64, orderNo, payNo, channel string, payAmount int64, paidAt int64) error {
	return tx.Model(&model.Order{}).
		Where("tenant_id = ? AND order_no = ? AND status = ?", tenantID, orderNo, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":      model.OrderStatusPaid,
			"pay_no":      payNo,
			"pay_channel": channel,
			"pay_amount":  payAmount,
			"paid_at":     paidAt,
		}).Error
}

func (r *OrderRepo) ListByUser(ctx context.Context, tenantID, userID uint64, page, pageSize int, status *int8, reviewPending bool) ([]model.Order, int64, error) {
	var list []model.Order
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Order{}).Where("tenant_id = ? AND user_id = ?", tenantID, userID)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if reviewPending {
		q = q.Where("status = ? AND review_status = ?", model.OrderStatusCompleted, model.ReviewStatusPending)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Items").Offset((page-1)*pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

func (r *OrderRepo) ListAdmin(ctx context.Context, tenantID uint64, page, pageSize int, status *int8) ([]model.Order, int64, error) {
	var list []model.Order
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Order{}).Where("tenant_id = ?", tenantID)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Items").Offset((page-1)*pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, tenantID uint64, orderNo string, fromStatus, toStatus int8, extra map[string]interface{}) error {
	updates := map[string]interface{}{"status": toStatus}
	for k, v := range extra {
		updates[k] = v
	}
	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("tenant_id = ? AND order_no = ? AND status = ?", tenantID, orderNo, fromStatus).
		Updates(updates).Error
}

func (r *OrderRepo) ListExpiredPending(ctx context.Context, now int64) ([]model.Order, error) {
	var list []model.Order
	err := r.db.WithContext(ctx).Preload("Items").
		Where("status = ? AND expire_at > 0 AND expire_at < ?", model.OrderStatusPending, now).
		Find(&list).Error
	return list, err
}

func (r *OrderRepo) CloseExpired(ctx context.Context, orderID uint64, now int64) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND status = ?", orderID, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":        model.OrderStatusClosed,
			"cancel_reason": "支付超时自动关闭",
			"cancelled_at":  now,
		}).Error
}

type PayNotifyRepo struct {
	db *gorm.DB
}

func NewPayNotifyRepo(db *gorm.DB) *PayNotifyRepo { return &PayNotifyRepo{db: db} }

func (r *PayNotifyRepo) Exists(ctx context.Context, notifyID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PayNotifyLog{}).Where("notify_id = ?", notifyID).Count(&count).Error
	return count > 0, err
}

func (r *PayNotifyRepo) Create(ctx context.Context, log *model.PayNotifyLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *PayNotifyRepo) MarkProcessed(ctx context.Context, notifyID string) error {
	return r.db.WithContext(ctx).Model(&model.PayNotifyLog{}).Where("notify_id = ?", notifyID).
		Update("processed", true).Error
}

func (r *OrderRepo) ListByOrderNos(ctx context.Context, tenantID uint64, orderNos []string) ([]model.Order, error) {
	if len(orderNos) == 0 {
		return nil, nil
	}
	var orders []model.Order
	err := r.db.WithContext(ctx).Preload("Items").
		Where("tenant_id = ? AND order_no IN ?", tenantID, orderNos).
		Find(&orders).Error
	return orders, err
}

func (r *OrderRepo) RunInTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *OrderRepo) Ship(ctx context.Context, tenantID uint64, orderNo, expressCompany, expressNo, shipRemark string, shippedAt int64) error {
	res := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("tenant_id = ? AND order_no = ? AND status = ?", tenantID, orderNo, model.OrderStatusPaid).
		Updates(map[string]interface{}{
			"status":           model.OrderStatusShipped,
			"shipped_at":       shippedAt,
			"express_company":  expressCompany,
			"express_no":       expressNo,
			"ship_remark":      shipRemark,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *OrderRepo) ConfirmReceive(ctx context.Context, tenantID, userID uint64, orderNo string, now int64) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("tenant_id = ? AND user_id = ? AND order_no = ? AND status = ?", tenantID, userID, orderNo, model.OrderStatusShipped).
		Updates(map[string]interface{}{
			"status":        model.OrderStatusCompleted,
			"received_at":   now,
			"completed_at":  now,
			"review_status": model.ReviewStatusPending,
		}).Error
}

func (r *OrderRepo) MarkReviewed(ctx context.Context, tenantID, userID uint64, orderNo string) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("tenant_id = ? AND user_id = ? AND order_no = ? AND review_status = ?", tenantID, userID, orderNo, model.ReviewStatusPending).
		Update("review_status", model.ReviewStatusDone).Error
}

func (r *OrderRepo) CancelPending(ctx context.Context, tenantID, userID uint64, orderNo, reason string, now int64) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("tenant_id = ? AND user_id = ? AND order_no = ? AND status = ?", tenantID, userID, orderNo, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":        model.OrderStatusCancelled,
			"cancel_reason": reason,
			"cancelled_at":  now,
		}).Error
}

func (r *OrderRepo) TouchExpireCheck(_ context.Context) {
	// placeholder for cron hook
	_ = time.Now()
}
