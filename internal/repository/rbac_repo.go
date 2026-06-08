package repository

import (
	"context"

	"sort"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type AdminRepo struct {
	db *gorm.DB
}

func NewAdminRepo(db *gorm.DB) *AdminRepo { return &AdminRepo{db: db} }

func (r *AdminRepo) FindByUsername(ctx context.Context, tenantID uint64, username string) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND username = ?", tenantID, username).First(&u).Error
	return &u, err
}

func (r *AdminRepo) FindByUsernameAll(ctx context.Context, username string) ([]model.AdminUser, error) {
	var list []model.AdminUser
	err := r.db.WithContext(ctx).Where("username = ? AND status = 1", username).Find(&list).Error
	return list, err
}

func (r *AdminRepo) FindByID(ctx context.Context, id uint64) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.WithContext(ctx).First(&u, id).Error
	return &u, err
}

func (r *AdminRepo) Create(ctx context.Context, u *model.AdminUser) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *AdminRepo) Update(ctx context.Context, u *model.AdminUser) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *AdminRepo) List(ctx context.Context, tenantID *uint64, page, pageSize int, keyword string) ([]model.AdminUser, int64, error) {
	var list []model.AdminUser
	var total int64
	q := r.db.WithContext(ctx).Model(&model.AdminUser{})
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

func (r *AdminRepo) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", id).Update("status", status).Error
}

type RBACRepo struct {
	db *gorm.DB
}

func NewRBACRepo(db *gorm.DB) *RBACRepo { return &RBACRepo{db: db} }

func (r *RBACRepo) ListMenus(ctx context.Context) ([]model.AdminMenu, error) {
	var list []model.AdminMenu
	err := r.db.WithContext(ctx).Where("status = 1").Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *RBACRepo) GetMenu(ctx context.Context, id uint64) (*model.AdminMenu, error) {
	var m model.AdminMenu
	err := r.db.WithContext(ctx).First(&m, id).Error
	return &m, err
}

func (r *RBACRepo) CreateMenu(ctx context.Context, m *model.AdminMenu) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *RBACRepo) UpdateMenu(ctx context.Context, m *model.AdminMenu) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *RBACRepo) DeleteMenu(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.AdminMenu{}, id).Error
}

func (r *RBACRepo) ListRoles(ctx context.Context, tenantID *uint64, scope string) ([]model.AdminRole, error) {
	var list []model.AdminRole
	q := r.db.WithContext(ctx).Where("status = 1")
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}
	if scope != "" {
		q = q.Where("scope = ?", scope)
	}
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}

func (r *RBACRepo) GetRoleByCode(ctx context.Context, tenantID uint64, code string) (*model.AdminRole, error) {
	var role model.AdminRole
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND code = ?", tenantID, code).First(&role).Error
	return &role, err
}

func (r *RBACRepo) GetRole(ctx context.Context, id uint64) (*model.AdminRole, error) {
	var role model.AdminRole
	err := r.db.WithContext(ctx).First(&role, id).Error
	return &role, err
}

func (r *RBACRepo) CreateRole(ctx context.Context, role *model.AdminRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *RBACRepo) UpdateRole(ctx context.Context, role *model.AdminRole) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *RBACRepo) DeleteRole(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.AdminRole{}, id).Error
}

func (r *RBACRepo) GetRoleMenuIDs(ctx context.Context, roleID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&model.AdminRoleMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &ids).Error
	return ids, err
}

func (r *RBACRepo) SetRoleMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.AdminRoleMenu{}).Error; err != nil {
			return err
		}
		for _, mid := range menuIDs {
			if err := tx.Create(&model.AdminRoleMenu{RoleID: roleID, MenuID: mid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RBACRepo) GetUserRoleIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&model.AdminUserRole{}).Where("user_id = ?", userID).Pluck("role_id", &ids).Error
	return ids, err
}

func (r *RBACRepo) SetUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.AdminUserRole{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := tx.Create(&model.AdminUserRole{UserID: userID, RoleID: rid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RBACRepo) GetUserRoleCodes(ctx context.Context, userID uint64) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Table("admin_user_roles ur").
		Joins("JOIN admin_roles r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.status = 1", userID).
		Pluck("r.code", &codes).Error
	return codes, err
}

func (r *RBACRepo) GetRoleIDsByCodes(ctx context.Context, codes []string) ([]uint64, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&model.AdminRole{}).
		Where("code IN ? AND status = 1", codes).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *RBACRepo) GetMenusByRoleIDs(ctx context.Context, roleIDs []uint64, scopes []string) ([]model.AdminMenu, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var list []model.AdminMenu
	q := r.db.WithContext(ctx).
		Table("admin_menus m").
		Joins("JOIN admin_role_menus rm ON rm.menu_id = m.id").
		Where("rm.role_id IN ? AND m.status = 1", roleIDs)
	if len(scopes) > 0 {
		q = q.Where("m.scope IN ?", scopes)
	}
	err := q.Group("m.id").Order("m.sort ASC, m.id ASC").Find(&list).Error
	return list, err
}

func (r *RBACRepo) ListAllAdminsWithRoles(ctx context.Context) ([]model.AdminUser, error) {
	var list []model.AdminUser
	err := r.db.WithContext(ctx).Where("status = 1").Find(&list).Error
	return list, err
}

func (r *RBACRepo) MenuCount(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.AdminMenu{}).Count(&n).Error
	return n, err
}

func (r *RBACRepo) RoleCount(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.AdminRole{}).Count(&n).Error
	return n, err
}

func (r *RBACRepo) RoleMenuCount(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.AdminRoleMenu{}).Count(&n).Error
	return n, err
}

func (r *RBACRepo) MenuIDsByPermissions(ctx context.Context, perms []string) ([]uint64, error) {
	if len(perms) == 0 {
		return nil, nil
	}
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&model.AdminMenu{}).
		Where("permission IN ? AND status = 1", perms).
		Order("sort ASC, id ASC").
		Pluck("id", &ids).Error
	return ids, err
}

func (r *RBACRepo) AllMenuIDs(ctx context.Context) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&model.AdminMenu{}).
		Where("status = 1").
		Order("sort ASC, id ASC").
		Pluck("id", &ids).Error
	return ids, err
}

func (r *RBACRepo) GetMenuByPermission(ctx context.Context, permission string) (*model.AdminMenu, error) {
	var m model.AdminMenu
	err := r.db.WithContext(ctx).Where("permission = ?", permission).First(&m).Error
	return &m, err
}

func (r *RBACRepo) UpdateMenuParent(ctx context.Context, id, parentID uint64) error {
	return r.db.WithContext(ctx).Model(&model.AdminMenu{}).Where("id = ?", id).
		Update("parent_id", parentID).Error
}

func (r *RBACRepo) ApplyMenuSpec(ctx context.Context, permission, path, scope string, parentID uint64, title string, sort int) error {
	q := r.db.WithContext(ctx).Model(&model.AdminMenu{}).Where("permission = ?", permission)
	if permission == "" {
		q = r.db.WithContext(ctx).Model(&model.AdminMenu{}).Where("path = ? AND scope = ?", path, scope)
	}
	return q.Updates(map[string]interface{}{
		"parent_id":  parentID,
		"permission": permission,
		"title":      title,
		"path":       path,
		"scope":      scope,
		"sort":       sort,
	}).Error
}

func (r *RBACRepo) DisableMenusExcept(ctx context.Context, keepPermissions []string) error {
	if len(keepPermissions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.AdminMenu{}).
		Where("permission NOT IN ?", keepPermissions).
		Update("status", 0).Error
}

func (r *RBACRepo) ExpandMenuAncestors(ctx context.Context, menus []model.AdminMenu) ([]model.AdminMenu, error) {
	if len(menus) == 0 {
		return menus, nil
	}
	all, err := r.ListMenus(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint64]model.AdminMenu, len(all))
	for _, m := range all {
		byID[m.ID] = m
	}
	seen := make(map[uint64]struct{}, len(menus))
	out := make([]model.AdminMenu, 0, len(menus))
	add := func(m model.AdminMenu) {
		if _, ok := seen[m.ID]; ok {
			return
		}
		seen[m.ID] = struct{}{}
		out = append(out, m)
	}
	var walkUp func(m model.AdminMenu)
	walkUp = func(m model.AdminMenu) {
		add(m)
		if m.ParentID == 0 {
			return
		}
		if parent, ok := byID[m.ParentID]; ok {
			walkUp(parent)
		}
	}
	for _, m := range menus {
		walkUp(m)
	}
	sortMenus(out)
	return out, nil
}

func sortMenus(list []model.AdminMenu) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Sort == list[j].Sort {
			return list[i].ID < list[j].ID
		}
		return list[i].Sort < list[j].Sort
	})
}
