package model

type FreightRule struct {
	BaseModel
	TenantScoped
	BaseFreight     int64  `gorm:"not null;default:800" json:"base_freight"`
	FreeThreshold   int64  `gorm:"not null;default:9900" json:"free_threshold"`
	RegionCode      string `gorm:"size:32;default:default" json:"region_code"`
	Status          int8   `gorm:"default:1" json:"status"`
}

func (FreightRule) TableName() string { return "freight_rules" }
