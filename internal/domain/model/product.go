package model

type ProductCategory struct {
	BaseModel
	TenantScoped
	ParentID uint64 `gorm:"default:0" json:"parent_id"`
	Name     string `gorm:"size:128;not null" json:"name"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   int8   `gorm:"default:1" json:"status"`
}

func (ProductCategory) TableName() string { return "product_categories" }

type Product struct {
	BaseModel
	TenantScoped
	CategoryID  uint64 `gorm:"index" json:"category_id"`
	Name        string `gorm:"size:256;not null" json:"name"`
	Subtitle    string `gorm:"size:512" json:"subtitle"`
	Description string `gorm:"type:text" json:"description"`
	CoverImage  string `gorm:"size:512" json:"cover_image"`
	Images      string `gorm:"type:text" json:"images"`
	Specs       string `gorm:"type:text" json:"specs"`
	Status      int8   `gorm:"default:1" json:"status"`
	SalesCount  int64  `gorm:"default:0" json:"sales_count"`
}

func (Product) TableName() string { return "products" }

type ProductSKU struct {
	BaseModel
	TenantScoped
	ProductID uint64 `gorm:"index;not null" json:"product_id"`
	SKUCode       string `gorm:"size:64;index" json:"sku_code"`
	Spec          string `gorm:"size:256" json:"spec"`
	SpecValues    string `gorm:"type:text" json:"spec_values"`
	Image         string `gorm:"size:512" json:"image"`
	Price         int64  `gorm:"not null" json:"price"`
	OriginalPrice int64  `gorm:"default:0" json:"original_price"`
	Stock         int64  `gorm:"not null;default:0" json:"stock"`
	Version   int64  `gorm:"default:0" json:"version"`
	Status    int8   `gorm:"default:1" json:"status"`
}

func (ProductSKU) TableName() string { return "product_skus" }
