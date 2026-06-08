package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type BalanceRepo struct {
	db *gorm.DB
}

func NewBalanceRepo(db *gorm.DB) *BalanceRepo { return &BalanceRepo{db: db} }

func (r *BalanceRepo) ChangeBalance(ctx context.Context, tenantID, userID uint64, change int64, log *model.UserBalanceLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.changeBalanceTx(tx, tenantID, userID, change, log)
	})
}

func (r *BalanceRepo) ChangeBalanceWithTx(tx *gorm.DB, tenantID, userID uint64, change int64, log *model.UserBalanceLog) error {
	return r.changeBalanceTx(tx, tenantID, userID, change, log)
}

func (r *BalanceRepo) changeBalanceTx(tx *gorm.DB, tenantID, userID uint64, change int64, log *model.UserBalanceLog) error {
	res := tx.Model(&model.User{}).
		Where("tenant_id = ? AND id = ? AND balance + ? >= 0", tenantID, userID, change).
		UpdateColumn("balance", gorm.Expr("balance + ?", change))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("balance not enough")
	}
	var u model.User
	if err := tx.Where("tenant_id = ? AND id = ?", tenantID, userID).First(&u).Error; err != nil {
		return err
	}
	log.BalanceAfter = u.Balance
	log.UserID = userID
	log.TenantID = tenantID
	return tx.Create(log).Error
}

func (r *BalanceRepo) ListLogs(ctx context.Context, tenantID, userID uint64, page, pageSize int) ([]model.UserBalanceLog, int64, error) {
	var list []model.UserBalanceLog
	var total int64
	q := r.db.WithContext(ctx).Model(&model.UserBalanceLog{}).Where("tenant_id = ? AND user_id = ?", tenantID, userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

type FreightRepo struct {
	db *gorm.DB
}

func NewFreightRepo(db *gorm.DB) *FreightRepo { return &FreightRepo{db: db} }

func (r *FreightRepo) GetDefault(ctx context.Context, tenantID uint64) (*model.FreightRule, error) {
	var rule model.FreightRule
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND status = 1", tenantID).Order("id ASC").First(&rule).Error
	return &rule, err
}

func (r *FreightRepo) UpsertDefault(ctx context.Context, rule *model.FreightRule) error {
	var existing model.FreightRule
	err := r.db.WithContext(ctx).Where("tenant_id = ?", rule.TenantID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(rule).Error
	}
	if err != nil {
		return err
	}
	existing.BaseFreight = rule.BaseFreight
	existing.FreeThreshold = rule.FreeThreshold
	return r.db.WithContext(ctx).Save(&existing).Error
}
