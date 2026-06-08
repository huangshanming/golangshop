package repository

import (
	"context"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type PlatformRepo struct {
	db *gorm.DB
}

func NewPlatformRepo(db *gorm.DB) *PlatformRepo { return &PlatformRepo{db: db} }

func (r *PlatformRepo) CountCategories(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PlatformCategory{}).Count(&count).Error
	return count, err
}

// --- categories ---

func (r *PlatformRepo) ListCategories(ctx context.Context, parentID uint64, limit int, onlyActive bool) ([]model.PlatformCategory, error) {
	var list []model.PlatformCategory
	q := r.db.WithContext(ctx).Model(&model.PlatformCategory{}).Where("parent_id = ?", parentID)
	if onlyActive {
		q = q.Where("status = 1")
	}
	q = q.Order("sort ASC, id ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&list).Error
	return list, err
}

func (r *PlatformRepo) ListAllCategories(ctx context.Context, onlyActive bool) ([]model.PlatformCategory, error) {
	var list []model.PlatformCategory
	q := r.db.WithContext(ctx).Model(&model.PlatformCategory{})
	if onlyActive {
		q = q.Where("status = 1")
	}
	err := q.Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *PlatformRepo) GetCategoryByID(ctx context.Context, id uint64) (*model.PlatformCategory, error) {
	var c model.PlatformCategory
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	return &c, err
}

func (r *PlatformRepo) CreateCategory(ctx context.Context, c *model.PlatformCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *PlatformRepo) UpdateCategory(ctx context.Context, c *model.PlatformCategory) error {
	return r.db.WithContext(ctx).Omit("CreatedAt").Save(c).Error
}

func (r *PlatformRepo) DeleteCategory(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.PlatformCategory{}, id).Error
}

func (r *PlatformRepo) CountChildCategories(ctx context.Context, parentID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PlatformCategory{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

func (r *PlatformRepo) CountProductsByCategory(ctx context.Context, categoryID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Product{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}

func (r *PlatformRepo) ListCategoryDescendantIDs(ctx context.Context, rootID uint64) ([]uint64, error) {
	if rootID == 0 {
		return nil, nil
	}
	all, err := r.ListAllCategories(ctx, false)
	if err != nil {
		return nil, err
	}
	children := make(map[uint64][]uint64)
	for _, c := range all {
		children[c.ParentID] = append(children[c.ParentID], c.ID)
	}
	ids := []uint64{rootID}
	queue := []uint64{rootID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range children[cur] {
			ids = append(ids, child)
			queue = append(queue, child)
		}
	}
	return ids, nil
}

func (r *PlatformRepo) GetCategoryNamesByIDs(ctx context.Context, ids []uint64) (map[uint64]string, error) {
	names := make(map[uint64]string)
	if len(ids) == 0 {
		return names, nil
	}
	var list []model.PlatformCategory
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
	if err != nil {
		return nil, err
	}
	for _, c := range list {
		names[c.ID] = c.Name
	}
	return names, nil
}

// --- banners ---

func (r *PlatformRepo) ListBanners(ctx context.Context, onlyActive bool, position string) ([]model.HomeBanner, error) {
	var list []model.HomeBanner
	q := r.db.WithContext(ctx).Model(&model.HomeBanner{})
	if onlyActive {
		q = q.Where("status = 1")
	}
	if position != "" {
		q = q.Where("position = ?", position)
	}
	err := q.Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *PlatformRepo) GetBannerByID(ctx context.Context, id uint64) (*model.HomeBanner, error) {
	var b model.HomeBanner
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&b).Error
	return &b, err
}

func (r *PlatformRepo) CreateBanner(ctx context.Context, b *model.HomeBanner) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *PlatformRepo) UpdateBanner(ctx context.Context, b *model.HomeBanner) error {
	return r.db.WithContext(ctx).Omit("CreatedAt").Save(b).Error
}

func (r *PlatformRepo) DeleteBanner(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.HomeBanner{}, id).Error
}

// --- settings ---

func (r *PlatformRepo) GetSetting(ctx context.Context, key string) (string, error) {
	var s model.PlatformSetting
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&s).Error
	if err != nil {
		return "", err
	}
	return s.Value, nil
}

func (r *PlatformRepo) UpsertSetting(ctx context.Context, key, value string) error {
	var s model.PlatformSetting
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(&model.PlatformSetting{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	s.Value = value
	return r.db.WithContext(ctx).Save(&s).Error
}

func (r *PlatformRepo) ListSettings(ctx context.Context) (map[string]string, error) {
	var list []model.PlatformSetting
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(list))
	for _, s := range list {
		out[s.Key] = s.Value
	}
	return out, nil
}

// --- featured ---

func (r *PlatformRepo) ListFeatured(ctx context.Context, onlyActive bool) ([]model.HomeFeaturedProduct, error) {
	var list []model.HomeFeaturedProduct
	q := r.db.WithContext(ctx).Model(&model.HomeFeaturedProduct{})
	if onlyActive {
		q = q.Where("status = 1")
	}
	err := q.Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *PlatformRepo) GetFeaturedByID(ctx context.Context, id uint64) (*model.HomeFeaturedProduct, error) {
	var f model.HomeFeaturedProduct
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&f).Error
	return &f, err
}

func (r *PlatformRepo) GetFeaturedByProductID(ctx context.Context, productID uint64) (*model.HomeFeaturedProduct, error) {
	var f model.HomeFeaturedProduct
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&f).Error
	return &f, err
}

func (r *PlatformRepo) CreateFeatured(ctx context.Context, f *model.HomeFeaturedProduct) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *PlatformRepo) UpdateFeatured(ctx context.Context, f *model.HomeFeaturedProduct) error {
	return r.db.WithContext(ctx).Omit("CreatedAt").Save(f).Error
}

func (r *PlatformRepo) DeleteFeatured(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.HomeFeaturedProduct{}, id).Error
}

func (r *PlatformRepo) UpdateFeaturedSorts(ctx context.Context, items []model.HomeFeaturedProduct) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range items {
			if err := tx.Model(&model.HomeFeaturedProduct{}).
				Where("id = ?", items[i].ID).
				Update("sort", items[i].Sort).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
