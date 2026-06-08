package model

type UserAddress struct {
	BaseModel
	TenantScoped
	UserID    uint64 `gorm:"index;not null" json:"user_id"`
	Name      string `gorm:"size:64;not null" json:"name"`
	Phone     string `gorm:"size:20;not null" json:"phone"`
	Province  string `gorm:"size:64" json:"province"`
	City      string `gorm:"size:64" json:"city"`
	District  string `gorm:"size:64" json:"district"`
	Detail    string `gorm:"size:512;not null" json:"detail"`
	IsDefault bool   `gorm:"default:false" json:"is_default"`
}

func (UserAddress) TableName() string { return "user_addresses" }
