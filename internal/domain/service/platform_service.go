package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
)

type PlatformService struct {
	repo        *repository.PlatformRepo
	productRepo *repository.ProductRepo
	tenantRepo  *repository.TenantRepo
}

func NewPlatformService(repo *repository.PlatformRepo, productRepo *repository.ProductRepo, tenantRepo *repository.TenantRepo) *PlatformService {
	return &PlatformService{repo: repo, productRepo: productRepo, tenantRepo: tenantRepo}
}

type PlatformCategoryReq struct {
	ParentID uint64 `json:"parent_id"`
	Name     string `json:"name" binding:"required"`
	Icon     string `json:"icon"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

type PlatformCategoryTreeItem struct {
	model.PlatformCategory
	Children []PlatformCategoryTreeItem `json:"children,omitempty"`
}

type BannerReq struct {
	Title      string `json:"title"`
	ImageURL   string `json:"image_url" binding:"required"`
	Position   string `json:"position"`
	Sort       int    `json:"sort"`
	Status     int8   `json:"status"`
	LinkType   string `json:"link_type" binding:"required"`
	LinkTarget string `json:"link_target"`
}

type SiteSettings struct {
	LogoURL  string `json:"logo_url"`
	SiteName string `json:"site_name"`
}

type FeaturedProductItem struct {
	ID          uint64 `json:"id"`
	ProductID   uint64 `json:"product_id"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
	ProductName string `json:"product_name"`
	TenantName  string `json:"tenant_name"`
	CoverImage  string `json:"cover_image"`
	MinPrice    int64  `json:"min_price"`
}

type AddFeaturedReq struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	Sort      int    `json:"sort"`
}

type FeaturedSortItem struct {
	ID   uint64 `json:"id" binding:"required"`
	Sort int    `json:"sort"`
}

func (s *PlatformService) EnsurePlatformCategories(ctx context.Context) error {
	count, err := s.repo.CountCategories(ctx)
	if err != nil {
		return err
	}
	if count == 0 {
		if err := s.seedDefaultCategories(ctx); err != nil {
			return err
		}
	}
	return s.migrateProductCategoryIDs(ctx)
}

func (s *PlatformService) seedDefaultCategories(ctx context.Context) error {
	specs := []struct {
		name     string
		sort     int
		children []struct {
			name string
			sort int
		}
	}{
		{name: "数码", sort: 1, children: []struct {
			name string
			sort int
		}{{name: "手机", sort: 1}, {name: "耳机", sort: 2}}},
		{name: "服饰", sort: 2, children: []struct {
			name string
			sort int
		}{{name: "男装", sort: 1}, {name: "女装", sort: 2}}},
		{name: "居家", sort: 3},
		{name: "美食", sort: 4},
	}
	for _, spec := range specs {
		parent := &model.PlatformCategory{Name: spec.name, Sort: spec.sort, Status: 1}
		if err := s.repo.CreateCategory(ctx, parent); err != nil {
			return err
		}
		for _, child := range spec.children {
			c := &model.PlatformCategory{ParentID: parent.ID, Name: child.name, Sort: child.sort, Status: 1}
			if err := s.repo.CreateCategory(ctx, c); err != nil {
				return err
			}
		}
	}
	return s.migrateProductCategoryIDs(ctx)
}

func (s *PlatformService) migrateProductCategoryIDs(ctx context.Context) error {
	type oldCat struct {
		ID       uint64
		TenantID uint64
		Name     string
	}
	var oldCats []oldCat
	if err := s.productRepo.DB().WithContext(ctx).
		Table("product_categories").
		Select("id, tenant_id, name").
		Find(&oldCats).Error; err != nil {
		return err
	}
	if len(oldCats) == 0 {
		return nil
	}
	platformCats, err := s.repo.ListAllCategories(ctx, false)
	if err != nil {
		return err
	}
	nameToID := make(map[string]uint64)
	for _, c := range platformCats {
		if c.ParentID == 0 {
			nameToID[c.Name] = c.ID
		}
	}
	for _, old := range oldCats {
		newID, ok := nameToID[old.Name]
		if !ok {
			continue
		}
		if err := s.productRepo.DB().WithContext(ctx).
			Model(&model.Product{}).
			Where("tenant_id = ? AND category_id = ?", old.TenantID, old.ID).
			Update("category_id", newID).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *PlatformService) ListCategoryTree(ctx context.Context, onlyActive bool) ([]PlatformCategoryTreeItem, error) {
	list, err := s.repo.ListAllCategories(ctx, onlyActive)
	if err != nil {
		return nil, err
	}
	return buildCategoryTree(list), nil
}

func buildCategoryTree(list []model.PlatformCategory) []PlatformCategoryTreeItem {
	byID := make(map[uint64]*PlatformCategoryTreeItem, len(list))
	for _, c := range list {
		cp := c
		byID[c.ID] = &PlatformCategoryTreeItem{PlatformCategory: cp}
	}
	var roots []PlatformCategoryTreeItem
	for _, node := range byID {
		if node.ParentID > 0 {
			if parent, ok := byID[node.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
				continue
			}
		}
		roots = append(roots, *node)
	}
	var sortTree func(items []PlatformCategoryTreeItem) []PlatformCategoryTreeItem
	sortTree = func(items []PlatformCategoryTreeItem) []PlatformCategoryTreeItem {
		for i := 0; i < len(items); i++ {
			for j := i + 1; j < len(items); j++ {
				if items[j].Sort < items[i].Sort || (items[j].Sort == items[i].Sort && items[j].ID < items[i].ID) {
					items[i], items[j] = items[j], items[i]
				}
			}
			if len(items[i].Children) > 0 {
				items[i].Children = sortTree(items[i].Children)
			}
		}
		return items
	}
	return sortTree(roots)
}

func (s *PlatformService) ListTopCategories(ctx context.Context, limit int) ([]model.PlatformCategory, error) {
	return s.repo.ListCategories(ctx, 0, limit, true)
}

func (s *PlatformService) ListCategoriesByParent(ctx context.Context, parentID uint64, limit int) ([]model.PlatformCategory, error) {
	return s.repo.ListCategories(ctx, parentID, limit, true)
}

func (s *PlatformService) CreateCategory(ctx context.Context, req PlatformCategoryReq) (*model.PlatformCategory, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("分类名称不能为空")
	}
	if req.ParentID > 0 {
		if _, err := s.repo.GetCategoryByID(ctx, req.ParentID); err != nil {
			return nil, fmt.Errorf("父分类不存在")
		}
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	c := &model.PlatformCategory{
		ParentID: req.ParentID,
		Name:     name,
		Icon:     strings.TrimSpace(req.Icon),
		Sort:     req.Sort,
		Status:   status,
	}
	return c, s.repo.CreateCategory(ctx, c)
}

func (s *PlatformService) UpdateCategory(ctx context.Context, id uint64, req PlatformCategoryReq) error {
	c, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return err
	}
	if req.ParentID == id {
		return errors.New("不能将分类设为自己的子分类")
	}
	if req.ParentID > 0 {
		if _, err := s.repo.GetCategoryByID(ctx, req.ParentID); err != nil {
			return fmt.Errorf("父分类不存在")
		}
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errors.New("分类名称不能为空")
	}
	c.ParentID = req.ParentID
	c.Name = name
	c.Icon = strings.TrimSpace(req.Icon)
	c.Sort = req.Sort
	if req.Status != 0 {
		c.Status = req.Status
	}
	return s.repo.UpdateCategory(ctx, c)
}

func (s *PlatformService) DeleteCategory(ctx context.Context, id uint64) error {
	childCount, err := s.repo.CountChildCategories(ctx, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("请先删除子分类")
	}
	prodCount, err := s.repo.CountProductsByCategory(ctx, id)
	if err != nil {
		return err
	}
	if prodCount > 0 {
		return errors.New("分类下仍有商品，无法删除")
	}
	return s.repo.DeleteCategory(ctx, id)
}

func (s *PlatformService) ValidateCategoryID(ctx context.Context, categoryID uint64) error {
	if categoryID == 0 {
		return errors.New("请选择商品分类")
	}
	c, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return errors.New("分类不存在")
	}
	if c.Status != 1 {
		return errors.New("分类已禁用")
	}
	return nil
}

func (s *PlatformService) ListBanners(ctx context.Context, onlyActive bool, position string) ([]model.HomeBanner, error) {
	return s.repo.ListBanners(ctx, onlyActive, position)
}

func (s *PlatformService) CreateBanner(ctx context.Context, req BannerReq) (*model.HomeBanner, error) {
	if err := validateBannerLink(req.LinkType, req.LinkTarget); err != nil {
		return nil, err
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	position := req.Position
	if position == "" {
		position = model.BannerPositionHome
	}
	b := &model.HomeBanner{
		Title: strings.TrimSpace(req.Title),
		ImageURL: strings.TrimSpace(req.ImageURL),
		Position: position,
		Sort: req.Sort,
		Status: status,
		LinkType: req.LinkType,
		LinkTarget: strings.TrimSpace(req.LinkTarget),
	}
	return b, s.repo.CreateBanner(ctx, b)
}

func (s *PlatformService) UpdateBanner(ctx context.Context, id uint64, req BannerReq) error {
	b, err := s.repo.GetBannerByID(ctx, id)
	if err != nil {
		return err
	}
	if err := validateBannerLink(req.LinkType, req.LinkTarget); err != nil {
		return err
	}
	b.Title = strings.TrimSpace(req.Title)
	b.ImageURL = strings.TrimSpace(req.ImageURL)
	if req.Position != "" {
		b.Position = req.Position
	}
	b.Sort = req.Sort
	if req.Status != 0 {
		b.Status = req.Status
	}
	b.LinkType = req.LinkType
	b.LinkTarget = strings.TrimSpace(req.LinkTarget)
	return s.repo.UpdateBanner(ctx, b)
}

func (s *PlatformService) DeleteBanner(ctx context.Context, id uint64) error {
	return s.repo.DeleteBanner(ctx, id)
}

func validateBannerLink(linkType, target string) error {
	switch linkType {
	case model.BannerLinkProduct, model.BannerLinkTenant, model.BannerLinkCategory:
		if target == "" {
			return fmt.Errorf("请填写跳转目标")
		}
		if _, err := strconv.ParseUint(target, 10, 64); err != nil {
			return fmt.Errorf("跳转目标 ID 无效")
		}
	case model.BannerLinkURL:
		if target == "" {
			return fmt.Errorf("请填写外链地址")
		}
	default:
		return fmt.Errorf("不支持的跳转类型")
	}
	return nil
}

func (s *PlatformService) GetSiteSettings(ctx context.Context) (SiteSettings, error) {
	m, err := s.repo.ListSettings(ctx)
	if err != nil {
		return SiteSettings{}, err
	}
	return SiteSettings{
		LogoURL:  m["logo_url"],
		SiteName: m["site_name"],
	}, nil
}

func (s *PlatformService) UpdateSiteSettings(ctx context.Context, req SiteSettings) error {
	if err := s.repo.UpsertSetting(ctx, "logo_url", strings.TrimSpace(req.LogoURL)); err != nil {
		return err
	}
	return s.repo.UpsertSetting(ctx, "site_name", strings.TrimSpace(req.SiteName))
}

func (s *PlatformService) EnsureDefaultSiteSettings(ctx context.Context) error {
	logo, err := s.repo.GetSetting(ctx, "logo_url")
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || logo == "" {
		_ = s.repo.UpsertSetting(ctx, "logo_url", "https://picsum.photos/seed/goshop-logo/200/200")
	}
	name, err := s.repo.GetSetting(ctx, "site_name")
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || name == "" {
		_ = s.repo.UpsertSetting(ctx, "site_name", "搜店网")
	} else if name == "GoShop" {
		_ = s.repo.UpsertSetting(ctx, "site_name", "搜店网")
	}
	return nil
}

func (s *PlatformService) ListFeaturedAdmin(ctx context.Context) ([]FeaturedProductItem, error) {
	list, err := s.repo.ListFeatured(ctx, false)
	if err != nil {
		return nil, err
	}
	return s.buildFeaturedItems(ctx, list)
}

func (s *PlatformService) ListFeaturedPublic(ctx context.Context) ([]FeaturedProductItem, error) {
	list, err := s.repo.ListFeatured(ctx, true)
	if err != nil {
		return nil, err
	}
	return s.buildFeaturedItems(ctx, list)
}

func (s *PlatformService) buildFeaturedItems(ctx context.Context, list []model.HomeFeaturedProduct) ([]FeaturedProductItem, error) {
	items := make([]FeaturedProductItem, 0, len(list))
	for _, f := range list {
		p, err := s.productRepo.GetByIDGlobal(ctx, f.ProductID)
		if err != nil {
			continue
		}
		if p.Status != 1 {
			continue
		}
		item := FeaturedProductItem{
			ID: f.ID, ProductID: f.ProductID, Sort: f.Sort, Status: f.Status,
			ProductName: p.Name, CoverImage: p.CoverImage,
		}
		if s.tenantRepo != nil {
			if tenant, err := s.tenantRepo.GetByID(ctx, p.TenantID); err == nil {
				item.TenantName = tenant.Name
			}
		}
		skus, _ := s.productRepo.ListSKUByProductsGlobal(ctx, []uint64{p.ID})
		for _, sku := range skus[p.ID] {
			if item.MinPrice == 0 || sku.Price < item.MinPrice {
				item.MinPrice = sku.Price
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *PlatformService) AddFeatured(ctx context.Context, req AddFeaturedReq) (*model.HomeFeaturedProduct, error) {
	p, err := s.productRepo.GetByIDGlobal(ctx, req.ProductID)
	if err != nil {
		return nil, errors.New("商品不存在")
	}
	if p.Status != 1 {
		return nil, errors.New("商品未上架")
	}
	if _, err := s.repo.GetFeaturedByProductID(ctx, req.ProductID); err == nil {
		return nil, errors.New("商品已在推荐列表中")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	f := &model.HomeFeaturedProduct{ProductID: req.ProductID, Sort: req.Sort, Status: 1}
	return f, s.repo.CreateFeatured(ctx, f)
}

func (s *PlatformService) DeleteFeatured(ctx context.Context, id uint64) error {
	return s.repo.DeleteFeatured(ctx, id)
}

func (s *PlatformService) UpdateFeaturedSort(ctx context.Context, items []FeaturedSortItem) error {
	updates := make([]model.HomeFeaturedProduct, 0, len(items))
	for _, it := range items {
		updates = append(updates, model.HomeFeaturedProduct{BaseModel: model.BaseModel{ID: it.ID}, Sort: it.Sort})
	}
	return s.repo.UpdateFeaturedSorts(ctx, updates)
}

func (s *PlatformService) EnsureDefaultBanners(ctx context.Context) error {
	homeBanners, err := s.repo.ListBanners(ctx, false, model.BannerPositionHome)
	if err != nil {
		return err
	}
	catBanners, err := s.repo.ListBanners(ctx, false, model.BannerPositionCategory)
	if err != nil {
		return err
	}
	cats, _ := s.repo.ListCategories(ctx, 0, 1, true)
	catID := uint64(0)
	if len(cats) > 0 {
		catID = cats[0].ID
	}
	var productID uint64
	s.productRepo.DB().WithContext(ctx).Model(&model.Product{}).Where("status = 1").Order("id ASC").Limit(1).Pluck("id", &productID)

	if len(homeBanners) == 0 {
		samples := []model.HomeBanner{
			{Title: "热门数码", ImageURL: "https://picsum.photos/seed/goshop-banner1/750/300", Position: model.BannerPositionHome, Sort: 1, Status: 1, LinkType: model.BannerLinkCategory, LinkTarget: fmt.Sprintf("%d", catID)},
			{Title: "精选好物", ImageURL: "https://picsum.photos/seed/goshop-banner2/750/300", Position: model.BannerPositionHome, Sort: 2, Status: 1, LinkType: model.BannerLinkProduct, LinkTarget: fmt.Sprintf("%d", productID)},
			{Title: "逛店铺", ImageURL: "https://picsum.photos/seed/goshop-banner3/750/300", Position: model.BannerPositionHome, Sort: 3, Status: 1, LinkType: model.BannerLinkTenant, LinkTarget: "1"},
		}
		for i := range samples {
			if samples[i].LinkTarget == "0" {
				continue
			}
			_ = s.repo.CreateBanner(ctx, &samples[i])
		}
	}
	if len(catBanners) == 0 && catID > 0 {
		_ = s.repo.CreateBanner(ctx, &model.HomeBanner{
			Title: "分类精选", ImageURL: "https://picsum.photos/seed/goshop-cat-banner/750/240",
			Position: model.BannerPositionCategory, Sort: 1, Status: 1,
			LinkType: model.BannerLinkCategory, LinkTarget: fmt.Sprintf("%d", catID),
		})
	}
	return nil
}

func (s *PlatformService) EnsureDefaultFeatured(ctx context.Context) error {
	list, err := s.repo.ListFeatured(ctx, false)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		return nil
	}
	var ids []uint64
	s.productRepo.DB().WithContext(ctx).Model(&model.Product{}).Where("status = 1").Order("id ASC").Limit(6).Pluck("id", &ids)
	for i, id := range ids {
		_ = s.repo.CreateFeatured(ctx, &model.HomeFeaturedProduct{ProductID: id, Sort: i + 1, Status: 1})
	}
	return nil
}
