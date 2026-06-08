package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"goshop/internal/config"
	"goshop/internal/domain/model"
	"goshop/internal/infra/database"
)

func main() {
	configPath := flag.String("config", "config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := database.NewMySQL(&cfg.Database)
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	tenant := ensureTenant(db)
	ensureAdmin(db, tenant.ID)
	ensureCategoriesAndProducts(db, tenant.ID)
	ensureFreightAndCoupons(db, tenant.ID)
	ensureDemoUser(db, tenant.ID)

	fmt.Println("seed completed: tenant_id=", tenant.ID)
}

func ensureTenant(db *gorm.DB) model.Tenant {
	var t model.Tenant
	if err := db.Where("code = ?", "default").First(&t).Error; err == nil {
		return t
	}
	t = model.Tenant{Name: "默认租户", Code: "default", Status: 1}
	if err := db.Create(&t).Error; err != nil {
		log.Fatalf("create tenant: %v", err)
	}
	return t
}

func ensureAdmin(db *gorm.DB, tenantID uint64) {
	var count int64
	db.Model(&model.AdminUser{}).Where("tenant_id = ? AND username = ?", tenantID, "admin").Count(&count)
	if count > 0 {
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.AdminUser{
		TenantScoped: model.TenantScoped{TenantID: tenantID},
		Username:     "admin",
		Password:     string(hash),
		Nickname:     "管理员",
		Role:         "admin",
		Status:       1,
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("create admin: %v", err)
	}
	fmt.Println("admin user created: admin / admin123")
}

type seedSpecDim struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type seedSKU struct {
	SpecValues    map[string]string
	Spec          string
	Price         int64
	OriginalPrice int64
	Stock         int64
}

type seedProduct struct {
	CategoryName string
	Name         string
	Subtitle     string
	Description  string
	CoverImage   string
	ExtraImages  []string
	SalesCount   int64
	Specs        []seedSpecDim
	SKUs         []seedSKU
}

func ensureCategoriesAndProducts(db *gorm.DB, tenantID uint64) {
	var count int64
	db.Model(&model.Product{}).Where("tenant_id = ?", tenantID).Count(&count)
	if count > 0 {
		return
	}

	categories := []model.ProductCategory{
		{TenantScoped: model.TenantScoped{TenantID: tenantID}, Name: "数码", Sort: 1, Status: 1},
		{TenantScoped: model.TenantScoped{TenantID: tenantID}, Name: "服饰", Sort: 2, Status: 1},
		{TenantScoped: model.TenantScoped{TenantID: tenantID}, Name: "居家", Sort: 3, Status: 1},
		{TenantScoped: model.TenantScoped{TenantID: tenantID}, Name: "美食", Sort: 4, Status: 1},
	}
	for i := range categories {
		if err := db.Create(&categories[i]).Error; err != nil {
			log.Fatalf("create category: %v", err)
		}
	}
	catMap := make(map[string]uint64)
	for _, c := range categories {
		catMap[c.Name] = c.ID
	}

	products := []seedProduct{
		{
			CategoryName: "数码",
			Name:         "无线降噪耳机 Pro",
			Subtitle:     "40dB 主动降噪 · 30 小时续航",
			Description:  "<p>采用定制驱动单元，支持蓝牙 5.3 与多设备切换。</p><p>适合通勤、办公与运动场景，配赠收纳盒与三尺寸耳塞。</p>",
			CoverImage:   "https://picsum.photos/seed/goshop-earphone/800/800",
			ExtraImages:  []string{"https://picsum.photos/seed/goshop-earphone2/800/800", "https://picsum.photos/seed/goshop-earphone3/800/800"},
			SalesCount:   1280,
			Specs:        []seedSpecDim{{Name: "颜色", Values: []string{"星空黑", "月光白", "限定蓝"}}},
			SKUs: []seedSKU{
				{SpecValues: map[string]string{"颜色": "星空黑"}, Spec: "星空黑", Price: 89900, OriginalPrice: 109900, Stock: 120},
				{SpecValues: map[string]string{"颜色": "月光白"}, Spec: "月光白", Price: 89900, OriginalPrice: 109900, Stock: 80},
				{SpecValues: map[string]string{"颜色": "限定蓝"}, Spec: "限定蓝", Price: 99900, OriginalPrice: 119900, Stock: 30},
			},
		},
		{
			CategoryName: "数码",
			Name:         "智能手表 S3",
			Subtitle:     "健康监测 · NFC 支付 · 5ATM 防水",
			Description:  "<p>1.43 英寸 AMOLED 屏幕，支持心率、血氧、睡眠监测。</p><p>内置 100+ 运动模式，续航可达 14 天。</p>",
			CoverImage:   "https://picsum.photos/seed/goshop-watch/800/800",
			ExtraImages:  []string{"https://picsum.photos/seed/goshop-watch2/800/800"},
			SalesCount:   860,
			Specs:        []seedSpecDim{{Name: "表带", Values: []string{"黑色", "银色"}}},
			SKUs: []seedSKU{
				{SpecValues: map[string]string{"表带": "黑色"}, Spec: "黑色", Price: 129900, OriginalPrice: 149900, Stock: 60},
				{SpecValues: map[string]string{"表带": "银色"}, Spec: "银色", Price: 129900, OriginalPrice: 149900, Stock: 45},
			},
		},
		{
			CategoryName: "服饰",
			Name:         "轻量防风夹克",
			Subtitle:     "春季新款 · 透气防泼水",
			Description:  "<p>采用高密度尼龙面料，轻便易收纳。</p><p>适合户外徒步与城市通勤，多口袋设计实用百搭。</p>",
			CoverImage:   "https://picsum.photos/seed/goshop-jacket/800/800",
			ExtraImages:  []string{"https://picsum.photos/seed/goshop-jacket2/800/800", "https://picsum.photos/seed/goshop-jacket3/800/800"},
			SalesCount:   520,
			Specs: []seedSpecDim{
				{Name: "尺码", Values: []string{"M", "L"}},
				{Name: "颜色", Values: []string{"深灰", "军绿"}},
			},
			SKUs: []seedSKU{
				{SpecValues: map[string]string{"尺码": "M", "颜色": "深灰"}, Spec: "M / 深灰", Price: 39900, OriginalPrice: 59900, Stock: 100},
				{SpecValues: map[string]string{"尺码": "L", "颜色": "深灰"}, Spec: "L / 深灰", Price: 39900, OriginalPrice: 59900, Stock: 80},
				{SpecValues: map[string]string{"尺码": "M", "颜色": "军绿"}, Spec: "M / 军绿", Price: 42900, OriginalPrice: 59900, Stock: 60},
				{SpecValues: map[string]string{"尺码": "L", "颜色": "军绿"}, Spec: "L / 军绿", Price: 42900, OriginalPrice: 59900, Stock: 50},
			},
		},
		{
			CategoryName: "服饰",
			Name:         "纯棉基础 T 恤",
			Subtitle:     "100% 精梳棉 · 多色可选",
			Description:  "<p>亲肤透气，版型修身不紧绷。</p><p>预缩水处理，机洗不易变形，日常百搭必备。</p>",
			CoverImage:   "https://picsum.photos/seed/goshop-tshirt/800/800",
			ExtraImages:  []string{"https://picsum.photos/seed/goshop-tshirt2/800/800"},
			SalesCount:   2100,
			Specs: []seedSpecDim{
				{Name: "颜色", Values: []string{"白色", "黑色"}},
				{Name: "尺码", Values: []string{"M", "L"}},
			},
			SKUs: []seedSKU{
				{SpecValues: map[string]string{"颜色": "白色", "尺码": "M"}, Spec: "白色 / M", Price: 7900, OriginalPrice: 12900, Stock: 200},
				{SpecValues: map[string]string{"颜色": "黑色", "尺码": "M"}, Spec: "黑色 / M", Price: 7900, OriginalPrice: 12900, Stock: 180},
				{SpecValues: map[string]string{"颜色": "白色", "尺码": "L"}, Spec: "白色 / L", Price: 7900, OriginalPrice: 12900, Stock: 150},
				{SpecValues: map[string]string{"颜色": "黑色", "尺码": "L"}, Spec: "黑色 / L", Price: 7900, OriginalPrice: 12900, Stock: 120},
			},
		},
		{
			CategoryName: "居家",
			Name:         "北欧简约台灯",
			Subtitle:     "三档调光 · 护眼无频闪",
			Description:  "<p>金属灯臂 + 磨砂灯罩，支持触控调光。</p><p>适合书桌、床头使用，营造温馨阅读氛围。</p>",
			CoverImage:   "https://picsum.photos/seed/goshop-lamp/800/800",
			ExtraImages:  []string{"https://picsum.photos/seed/goshop-lamp2/800/800"},
			SalesCount:   340,
			Specs:        []seedSpecDim{{Name: "颜色", Values: []string{"暖白色", "黑色"}}},
			SKUs: []seedSKU{
				{SpecValues: map[string]string{"颜色": "暖白色"}, Spec: "暖白色", Price: 15900, OriginalPrice: 19900, Stock: 90},
				{SpecValues: map[string]string{"颜色": "黑色"}, Spec: "黑色", Price: 16900, OriginalPrice: 20900, Stock: 70},
			},
		},
		{
			CategoryName: "美食",
			Name:         "精品挂耳咖啡礼盒",
			Subtitle:     "12 种风味 · 新鲜烘焙",
			Description:  "<p>精选阿拉比卡豆，独立包装锁鲜。</p><p>包含浅烘、中烘、深烘多种风味，送礼自饮皆宜。</p>",
			CoverImage:   "https://picsum.photos/seed/goshop-coffee/800/800",
			ExtraImages:  []string{"https://picsum.photos/seed/goshop-coffee2/800/800", "https://picsum.photos/seed/goshop-coffee3/800/800"},
			SalesCount:   680,
			Specs:        []seedSpecDim{{Name: "款式", Values: []string{"经典款 12 包", "限定款 12 包"}}},
			SKUs: []seedSKU{
				{SpecValues: map[string]string{"款式": "经典款 12 包"}, Spec: "经典款 12 包", Price: 8900, OriginalPrice: 12900, Stock: 300},
				{SpecValues: map[string]string{"款式": "限定款 12 包"}, Spec: "限定款 12 包", Price: 11900, OriginalPrice: 15900, Stock: 150},
			},
		},
	}

	for _, sp := range products {
		images, _ := json.Marshal(append([]string{sp.CoverImage}, sp.ExtraImages...))
		specsJSON, _ := json.Marshal(sp.Specs)
		p := model.Product{
			TenantScoped: model.TenantScoped{TenantID: tenantID},
			CategoryID:   catMap[sp.CategoryName],
			Name:         sp.Name,
			Subtitle:     sp.Subtitle,
			Description:  sp.Description,
			CoverImage:   sp.CoverImage,
			Images:       string(images),
			Specs:        string(specsJSON),
			Status:       1,
			SalesCount:   sp.SalesCount,
		}
		if err := db.Create(&p).Error; err != nil {
			log.Fatalf("create product: %v", err)
		}
		for _, s := range sp.SKUs {
			specValuesJSON, _ := json.Marshal(s.SpecValues)
			sku := model.ProductSKU{
				TenantScoped:  model.TenantScoped{TenantID: tenantID},
				ProductID:     p.ID,
				Spec:          s.Spec,
				SpecValues:    string(specValuesJSON),
				Price:         s.Price,
				OriginalPrice: s.OriginalPrice,
				Stock:         s.Stock,
				Status:        1,
			}
			if err := db.Create(&sku).Error; err != nil {
				log.Fatalf("create sku: %v", err)
			}
		}
	}
	fmt.Println("demo categories and products created")
}

func ensureFreightAndCoupons(db *gorm.DB, tenantID uint64) {
	var frCount int64
	db.Model(&model.FreightRule{}).Where("tenant_id = ?", tenantID).Count(&frCount)
	if frCount == 0 {
		_ = db.Create(&model.FreightRule{
			TenantScoped: model.TenantScoped{TenantID: tenantID},
			BaseFreight: 800, FreeThreshold: 9900, RegionCode: "default", Status: 1,
		}).Error
	}
	var cCount int64
	db.Model(&model.Coupon{}).Where("tenant_id = ?", tenantID).Count(&cCount)
	if cCount > 0 {
		return
	}
	now := time.Now().Unix()
	coupons := []model.Coupon{
		{
			TenantScoped: model.TenantScoped{TenantID: tenantID},
			Name: "满99减10", Type: model.CouponTypeFixed,
			ThresholdAmount: 9900, DiscountValue: 1000,
			ValidFrom: now - 86400, ValidTo: now + 86400*365,
			TotalCount: 1000, PerUserLimit: 3, Status: 1,
		},
		{
			TenantScoped: model.TenantScoped{TenantID: tenantID},
			Name: "满199打9折", Type: model.CouponTypePercent,
			ThresholdAmount: 19900, DiscountValue: 10, MaxDiscount: 5000,
			ValidFrom: now - 86400, ValidTo: now + 86400*365,
			TotalCount: 500, PerUserLimit: 2, Status: 1,
		},
	}
	for i := range coupons {
		if err := db.Create(&coupons[i]).Error; err != nil {
			log.Fatalf("create coupon: %v", err)
		}
	}
	fmt.Println("freight rule and coupons created")
}

func ensureDemoUser(db *gorm.DB, tenantID uint64) {
	var u model.User
	if err := db.Where("tenant_id = ? AND phone = ?", tenantID, "13800138000").First(&u).Error; err != nil {
		u = model.User{
			TenantScoped: model.TenantScoped{TenantID: tenantID},
			Phone: "13800138000", Nickname: "演示用户", Balance: 50000, Status: 1,
		}
		if err := db.Create(&u).Error; err != nil {
			log.Fatalf("create demo user: %v", err)
		}
	} else if u.Balance == 0 {
		_ = db.Model(&u).Update("balance", 50000).Error
	}
	var addrCount int64
	db.Model(&model.UserAddress{}).Where("tenant_id = ? AND user_id = ?", tenantID, u.ID).Count(&addrCount)
	if addrCount == 0 {
		_ = db.Create(&model.UserAddress{
			TenantScoped: model.TenantScoped{TenantID: tenantID},
			UserID: u.ID, Name: "张三", Phone: "13800138000",
			Province: "广东省", City: "深圳市", District: "南山区",
			Detail: "科技园南路 88 号", IsDefault: true,
		}).Error
	}
	var coupons []model.Coupon
	db.Where("tenant_id = ?", tenantID).Find(&coupons)
	for _, c := range coupons {
		var ucCount int64
		db.Model(&model.UserCoupon{}).Where("tenant_id = ? AND user_id = ? AND coupon_id = ?", tenantID, u.ID, c.ID).Count(&ucCount)
		if ucCount == 0 {
			_ = db.Create(&model.UserCoupon{
				TenantScoped: model.TenantScoped{TenantID: tenantID},
				UserID: u.ID, CouponID: c.ID, Status: model.UserCouponUnused, ReceivedAt: time.Now().Unix(),
			}).Error
		}
	}
	fmt.Println("demo user ready: phone 13800138000, balance 500.00 yuan")
}
