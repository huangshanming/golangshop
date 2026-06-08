package model

type CartItem struct {
	BaseModel
	TenantScoped
	UserID    uint64 `gorm:"index;not null" json:"user_id"`
	ProductID uint64 `gorm:"not null" json:"product_id"`
	SkuID     uint64 `gorm:"column:sku_id;not null" json:"sku_id"`
	Quantity  int    `gorm:"not null;default:1" json:"quantity"`
	Checked   bool   `gorm:"default:true" json:"checked"`
}

func (CartItem) TableName() string { return "cart_items" }
