package model

const (
	BannerLinkProduct  = "product"
	BannerLinkTenant   = "tenant"
	BannerLinkCategory = "category"
	BannerLinkURL      = "url"

	BannerPositionHome     = "home"
	BannerPositionCategory = "category"
)

type PlatformCategory struct {
	BaseModel
	ParentID uint64 `gorm:"default:0;index" json:"parent_id"`
	Name     string `gorm:"size:128;not null" json:"name"`
	Icon     string `gorm:"size:512" json:"icon"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   int8   `gorm:"default:1" json:"status"`
}

func (PlatformCategory) TableName() string { return "platform_categories" }

type HomeBanner struct {
	BaseModel
	Title      string `gorm:"size:128" json:"title"`
	ImageURL   string `gorm:"size:512;not null" json:"image_url"`
	Position   string `gorm:"size:32;default:home;index" json:"position"`
	Sort       int    `gorm:"default:0" json:"sort"`
	Status     int8   `gorm:"default:1" json:"status"`
	LinkType   string `gorm:"size:32;not null" json:"link_type"`
	LinkTarget string `gorm:"size:512" json:"link_target"`
}

func (HomeBanner) TableName() string { return "home_banners" }

type PlatformSetting struct {
	BaseModel
	Key   string `gorm:"size:64;uniqueIndex;not null" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func (PlatformSetting) TableName() string { return "platform_settings" }

type HomeFeaturedProduct struct {
	BaseModel
	ProductID uint64 `gorm:"uniqueIndex;not null" json:"product_id"`
	Sort      int    `gorm:"default:0" json:"sort"`
	Status    int8   `gorm:"default:1" json:"status"`
}

func (HomeFeaturedProduct) TableName() string { return "home_featured_products" }
