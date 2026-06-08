package model

const (
	AdminUserTypePlatform = "platform"
	AdminUserTypeTenant   = "tenant"

	MenuScopePlatform = "platform"
	MenuScopeTenant   = "tenant"
	MenuScopeAll      = "all"

	RoleScopePlatform = "platform"
	RoleScopeTenant   = "tenant"
)

type AdminMenu struct {
	BaseModel
	ParentID   uint64 `gorm:"index;default:0" json:"parent_id"`
	Title      string `gorm:"size:64;not null" json:"title"`
	Path       string `gorm:"size:128" json:"path"`
	Icon       string `gorm:"size:64" json:"icon"`
	Sort       int    `gorm:"default:0" json:"sort"`
	Scope      string `gorm:"size:16;default:tenant;index" json:"scope"`
	Permission string `gorm:"size:64;index" json:"permission"`
	Status     int8   `gorm:"default:1" json:"status"`
}

func (AdminMenu) TableName() string { return "admin_menus" }

type AdminRole struct {
	BaseModel
	TenantID uint64 `gorm:"uniqueIndex:idx_role_tenant_code,priority:1;index;not null" json:"tenant_id"`
	Code     string `gorm:"size:32;uniqueIndex:idx_role_tenant_code,priority:2;not null" json:"code"`
	Name     string `gorm:"size:64;not null" json:"name"`
	Scope    string `gorm:"size:16;default:tenant;index" json:"scope"`
	Status   int8   `gorm:"default:1" json:"status"`
}

func (AdminRole) TableName() string { return "admin_roles" }

type AdminRoleMenu struct {
	RoleID uint64 `gorm:"primaryKey" json:"role_id"`
	MenuID uint64 `gorm:"primaryKey" json:"menu_id"`
}

func (AdminRoleMenu) TableName() string { return "admin_role_menus" }

type AdminUserRole struct {
	UserID uint64 `gorm:"primaryKey" json:"user_id"`
	RoleID uint64 `gorm:"primaryKey" json:"role_id"`
}

func (AdminUserRole) TableName() string { return "admin_user_roles" }
