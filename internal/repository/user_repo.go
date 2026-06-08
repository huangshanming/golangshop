package repository

import (
	"context"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByPhone(ctx context.Context, tenantID uint64, phone string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND phone = ?", tenantID, phone).First(&u).Error
	return &u, err
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepo) FindByID(ctx context.Context, tenantID, id uint64) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&u).Error
	return &u, err
}

func (r *UserRepo) List(ctx context.Context, tenantID uint64, page, pageSize int, keyword string) ([]model.User, int64, error) {
	var list []model.User
	var total int64
	q := r.db.WithContext(ctx).Model(&model.User{}).Where("tenant_id = ?", tenantID)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("phone LIKE ? OR nickname LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

func (r *UserRepo) UpdateProfile(ctx context.Context, tenantID, userID uint64, nickname string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, userID).
		Update("nickname", nickname).Error
}
