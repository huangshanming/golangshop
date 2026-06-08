package auth

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

func TestSuperAdminAuthMe(t *testing.T) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		t.Fatal(err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = e.AddPolicy("super_admin", "/api/v1/admin/*", "*", "0")
	_, _ = e.AddPolicy("super_admin", "/api/v1/admin/*", "*", "*")
	_, _ = e.AddGroupingPolicy("superadmin", "super_admin")

	ok, err := e.Enforce("superadmin", "/api/v1/admin/auth/me", "GET", "0")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("superadmin should access /auth/me in platform view")
	}
}

func TestTenantAdminIsolation(t *testing.T) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		t.Fatal(err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = e.AddPolicy("tenant_admin", "/api/v1/admin/*", "*", "1")
	_, _ = e.AddGroupingPolicy("admin", "tenant_admin")

	ok, _ := e.Enforce("admin", "/api/v1/admin/products", "GET", "1")
	if !ok {
		t.Fatal("tenant admin should access tenant 1")
	}
	ok, _ = e.Enforce("admin", "/api/v1/admin/products", "GET", "2")
	if ok {
		t.Fatal("tenant admin should not access tenant 2 with wildcard dom policy")
	}
}
