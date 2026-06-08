package auth

import (
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	domainmodel "goshop/internal/domain/model"
)

const modelConf = `
[request_definition]
r = sub, obj, act, dom

[policy_definition]
p = sub, obj, act, dom

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*") && (r.dom == p.dom || p.dom == "*")
`

const adminAPIPath = "/api/v1/admin/*"

func NewEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, err
	}
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}
	if err := SyncPolicies(enforcer, db); err != nil {
		return nil, err
	}
	return enforcer, nil
}

func SyncPolicies(enforcer *casbin.Enforcer, db *gorm.DB) error {
	enforcer.ClearPolicy()

	var roles []domainmodel.AdminRole
	if err := db.Where("status = 1").Find(&roles).Error; err != nil {
		return err
	}
	roleByID := make(map[uint64]domainmodel.AdminRole, len(roles))
	var tenantIDs []uint64
	_ = db.Model(&domainmodel.Tenant{}).Pluck("id", &tenantIDs).Error

	for _, r := range roles {
		roleByID[r.ID] = r
		switch {
		case r.Code == "super_admin":
			_, _ = enforcer.AddPolicy(r.Code, adminAPIPath, "*", "0")
			_, _ = enforcer.AddPolicy(r.Code, adminAPIPath, "*", "*")
		case r.Scope == domainmodel.RoleScopeTenant:
			for _, tid := range tenantIDs {
				_, _ = enforcer.AddPolicy(r.Code, adminAPIPath, "*", strconv.FormatUint(tid, 10))
			}
		default:
			dom := roleDomain(r)
			_, _ = enforcer.AddPolicy(r.Code, adminAPIPath, "*", dom)
		}
	}

	var admins []domainmodel.AdminUser
	if err := db.Where("status = 1").Find(&admins).Error; err != nil {
		return err
	}

	for _, admin := range admins {
		var roleIDs []uint64
		db.Model(&domainmodel.AdminUserRole{}).Where("user_id = ?", admin.ID).Pluck("role_id", &roleIDs)

		codes := make([]string, 0)
		for _, rid := range roleIDs {
			if role, ok := roleByID[rid]; ok {
				codes = append(codes, role.Code)
			}
		}
		if len(codes) == 0 && admin.Role != "" {
			codes = append(codes, admin.Role)
		}

		for _, code := range codes {
			_, _ = enforcer.AddGroupingPolicy(admin.Username, code)
		}
	}

	return enforcer.SavePolicy()
}

func roleDomain(role domainmodel.AdminRole) string {
	if role.Scope == domainmodel.RoleScopePlatform {
		return "0"
	}
	if role.TenantID > 0 {
		return strconv.FormatUint(role.TenantID, 10)
	}
	return "*"
}

// EnsureTenantPolicies kept for backward compatibility during migration.
func EnsureTenantPolicies(enforcer *casbin.Enforcer, db *gorm.DB, tenantID uint64) error {
	return SyncPolicies(enforcer, db)
}

func SyncAdminPolicies(enforcer *casbin.Enforcer, db *gorm.DB) error {
	return SyncPolicies(enforcer, db)
}
