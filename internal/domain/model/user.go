package model

type User struct {
	BaseModel
	TenantScoped
	OpenID   string `gorm:"size:128;index" json:"openid"`
	Phone    string `gorm:"size:20;index" json:"phone"`
	Nickname string `gorm:"size:64" json:"nickname"`
	Avatar   string `gorm:"size:512" json:"avatar"`
	Balance  int64  `gorm:"default:0" json:"balance"`
	Status   int8   `gorm:"default:1" json:"status"`
}

func (User) TableName() string { return "users" }

type AdminUser struct {
	BaseModel
	TenantID uint64 `gorm:"uniqueIndex:idx_tenant_username,priority:1;index;not null" json:"tenant_id"`
	Username string `gorm:"size:64;uniqueIndex:idx_tenant_username,priority:2;not null" json:"username"`
	Password string `gorm:"size:128;not null" json:"-"`
	Nickname string `gorm:"size:64" json:"nickname"`
	UserType string `gorm:"size:16;default:tenant;index" json:"user_type"`
	Role     string `gorm:"size:32;default:admin" json:"role"`
	Status   int8   `gorm:"default:1" json:"status"`
}

func (AdminUser) TableName() string { return "admin_users" }
