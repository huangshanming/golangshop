package model

type Tenant struct {
	BaseModel
	Name   string `gorm:"size:128;not null" json:"name"`
	Code   string `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Status int8   `gorm:"default:1" json:"status"`
}

func (Tenant) TableName() string { return "tenants" }
