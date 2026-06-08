package service

import (
	"context"
	"errors"
	"sort"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
	"goshop/pkg/jwtutil"
)

type AdminLoginCandidate struct {
	UserID     uint64 `json:"user_id"`
	TenantID   uint64 `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
}

type AdminLoginResult struct {
	Token          string                `json:"token,omitempty"`
	Admin          *model.AdminUser      `json:"admin,omitempty"`
	NeedTenantPick bool                  `json:"need_tenant_pick,omitempty"`
	Candidates     []AdminLoginCandidate `json:"candidates,omitempty"`
}

type MenuTreeNode struct {
	model.AdminMenu
	Children []MenuTreeNode `json:"children,omitempty"`
}

type AdminMeResult struct {
	Admin        *model.AdminUser `json:"admin"`
	Roles        []string         `json:"roles"`
	Menus        []MenuTreeNode   `json:"menus"`
	ActiveTenant *model.Tenant    `json:"active_tenant,omitempty"`
	PlatformView bool             `json:"platform_view"`
}

type RBACService struct {
	rbacRepo   *repository.RBACRepo
	adminRepo  *repository.AdminRepo
	tenantRepo *repository.TenantRepo
	jwtMgr     *jwtutil.Manager
}

func NewRBACService(rbacRepo *repository.RBACRepo, adminRepo *repository.AdminRepo, tenantRepo *repository.TenantRepo, jwtMgr *jwtutil.Manager) *RBACService {
	return &RBACService{rbacRepo: rbacRepo, adminRepo: adminRepo, tenantRepo: tenantRepo, jwtMgr: jwtMgr}
}

func (s *RBACService) AdminLogin(ctx context.Context, username, password string, tenantID *uint64) (*AdminLoginResult, error) {
	var admins []model.AdminUser
	var err error
	if tenantID != nil {
		admin, err := s.adminRepo.FindByUsername(ctx, *tenantID, username)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		if err != nil {
			return nil, err
		}
		admins = []model.AdminUser{*admin}
	} else {
		admins, err = s.adminRepo.FindByUsernameAll(ctx, username)
		if err != nil {
			return nil, err
		}
	}

	matched := make([]model.AdminUser, 0, len(admins))
	for _, a := range admins {
		if a.Status != 1 {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password)) == nil {
			matched = append(matched, a)
		}
	}
	if len(matched) == 0 {
		return nil, errors.New("用户名或密码错误")
	}
	if len(matched) > 1 {
		candidates := make([]AdminLoginCandidate, 0, len(matched))
		for _, a := range matched {
			name := "平台"
			if a.TenantID > 0 {
				if t, err := s.tenantRepo.GetByID(ctx, a.TenantID); err == nil {
					name = t.Name
				}
			}
			candidates = append(candidates, AdminLoginCandidate{
				UserID: a.ID, TenantID: a.TenantID, TenantName: name,
			})
		}
		return &AdminLoginResult{NeedTenantPick: true, Candidates: candidates}, nil
	}

	token, admin, err := s.issueToken(ctx, &matched[0], initialActiveTenant(&matched[0]))
	if err != nil {
		return nil, err
	}
	return &AdminLoginResult{Token: token, Admin: admin}, nil
}

func initialActiveTenant(admin *model.AdminUser) uint64 {
	userType := adminUserType(admin)
	if userType == model.AdminUserTypeTenant {
		return admin.TenantID
	}
	return 0
}

func adminUserType(admin *model.AdminUser) string {
	if admin.UserType != "" {
		return admin.UserType
	}
	if admin.TenantID == 0 {
		return model.AdminUserTypePlatform
	}
	return model.AdminUserTypeTenant
}

func (s *RBACService) issueToken(ctx context.Context, admin *model.AdminUser, activeTenantID uint64) (string, *model.AdminUser, error) {
	roles, err := s.rbacRepo.GetUserRoleCodes(ctx, admin.ID)
	if err != nil {
		return "", nil, err
	}
	if len(roles) == 0 && admin.Role != "" {
		roles = []string{admin.Role}
	}
	userType := adminUserType(admin)
	if userType == model.AdminUserTypeTenant {
		activeTenantID = admin.TenantID
	}
	token, err := s.jwtMgr.GenerateAdminToken(jwtutil.AdminTokenParams{
		UserID: admin.ID, TenantID: admin.TenantID, ActiveTenantID: activeTenantID,
		UserType: userType, Roles: roles, Username: admin.Username,
	})
	return token, admin, err
}

func (s *RBACService) EnterTenant(ctx context.Context, adminID, tenantID uint64) (string, *model.AdminUser, error) {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return "", nil, err
	}
	if adminUserType(admin) != model.AdminUserTypePlatform {
		return "", nil, errors.New("仅平台管理员可进入租户")
	}
	if _, err := s.tenantRepo.GetByID(ctx, tenantID); err != nil {
		return "", nil, errors.New("租户不存在")
	}
	return s.issueToken(ctx, admin, tenantID)
}

func (s *RBACService) ExitTenant(ctx context.Context, adminID uint64) (string, *model.AdminUser, error) {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return "", nil, err
	}
	if adminUserType(admin) != model.AdminUserTypePlatform {
		return "", nil, errors.New("仅平台管理员可退出租户视角")
	}
	return s.issueToken(ctx, admin, 0)
}

func (s *RBACService) GetMe(ctx context.Context, adminID uint64, userType string, activeTenantID uint64) (*AdminMeResult, error) {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	roles, err := s.rbacRepo.GetUserRoleCodes(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 && admin.Role != "" {
		roles = []string{admin.Role}
	}
	platformView := userType == model.AdminUserTypePlatform && activeTenantID == 0
	roleIDs, err := s.rbacRepo.GetUserRoleIDs(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 && len(roles) > 0 {
		roleIDs, _ = s.rbacRepo.GetRoleIDsByCodes(ctx, roles)
	}
	menus, err := s.rbacRepo.GetMenusByRoleIDs(ctx, roleIDs, menuScopes(platformView))
	if err != nil {
		return nil, err
	}
	menus, err = s.rbacRepo.ExpandMenuAncestors(ctx, menus)
	if err != nil {
		return nil, err
	}
	res := &AdminMeResult{
		Admin: admin, Roles: roles, Menus: buildMenuTree(menus), PlatformView: platformView,
	}
	if activeTenantID > 0 {
		if t, err := s.tenantRepo.GetByID(ctx, activeTenantID); err == nil {
			res.ActiveTenant = t
		}
	}
	return res, nil
}

func menuScopes(platformView bool) []string {
	if platformView {
		return []string{model.MenuScopePlatform, model.MenuScopeAll}
	}
	return []string{model.MenuScopeTenant, model.MenuScopeAll}
}

func buildMenuTree(menus []model.AdminMenu) []MenuTreeNode {
	if len(menus) == 0 {
		return []MenuTreeNode{}
	}
	type node struct {
		model.AdminMenu
		children []*node
	}
	byID := make(map[uint64]*node, len(menus))
	for _, m := range menus {
		copyM := m
		byID[m.ID] = &node{AdminMenu: copyM}
	}
	var roots []*node
	for _, m := range menus {
		n := byID[m.ID]
		if m.ParentID == 0 || byID[m.ParentID] == nil {
			roots = append(roots, n)
			continue
		}
		byID[m.ParentID].children = append(byID[m.ParentID].children, n)
	}
	var convert func(n *node) MenuTreeNode
	convert = func(n *node) MenuTreeNode {
		out := MenuTreeNode{AdminMenu: n.AdminMenu, Children: []MenuTreeNode{}}
		sort.Slice(n.children, func(i, j int) bool {
			if n.children[i].Sort == n.children[j].Sort {
				return n.children[i].ID < n.children[j].ID
			}
			return n.children[i].Sort < n.children[j].Sort
		})
		for _, c := range n.children {
			out.Children = append(out.Children, convert(c))
		}
		return out
	}
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].Sort == roots[j].Sort {
			return roots[i].ID < roots[j].ID
		}
		return roots[i].Sort < roots[j].Sort
	})
	out := make([]MenuTreeNode, len(roots))
	for i, r := range roots {
		out[i] = convert(r)
	}
	return out
}

type defaultMenuSpec struct {
	Permission string
	ParentPerm string
	Title      string
	Path       string
	Icon       string
	Sort       int
	Scope      string
}

func defaultMenuSpecs() []defaultMenuSpec {
	return []defaultMenuSpec{
		{Permission: "dashboard:view", Title: "仪表盘", Path: "/dashboard", Icon: "Odometer", Sort: 1, Scope: model.MenuScopeAll},
		{Permission: "platform:group", Title: "平台管理", Path: "", Icon: "Setting", Sort: 10, Scope: model.MenuScopePlatform},
		{Permission: "tenant:list", ParentPerm: "platform:group", Title: "租户管理", Path: "/tenants", Icon: "OfficeBuilding", Sort: 11, Scope: model.MenuScopePlatform},
		{Permission: "menu:list", ParentPerm: "platform:group", Title: "菜单管理", Path: "/menus", Icon: "Menu", Sort: 12, Scope: model.MenuScopePlatform},
		{Permission: "role:list", ParentPerm: "platform:group", Title: "角色管理", Path: "/roles", Icon: "UserFilled", Sort: 13, Scope: model.MenuScopePlatform},
		{Permission: "admin:platform:list", ParentPerm: "platform:group", Title: "平台账号", Path: "/admins", Icon: "Avatar", Sort: 14, Scope: model.MenuScopePlatform},
		{Permission: "category:platform:list", ParentPerm: "platform:group", Title: "分类管理", Path: "/categories", Icon: "Grid", Sort: 15, Scope: model.MenuScopePlatform},
		{Permission: "banner:list", ParentPerm: "platform:group", Title: "Banner管理", Path: "/banners", Icon: "Picture", Sort: 16, Scope: model.MenuScopePlatform},
		{Permission: "site:settings", ParentPerm: "platform:group", Title: "站点设置", Path: "/site-settings", Icon: "Setting", Sort: 17, Scope: model.MenuScopePlatform},
		{Permission: "home:featured:list", ParentPerm: "platform:group", Title: "首页推荐", Path: "/home-featured", Icon: "Star", Sort: 18, Scope: model.MenuScopePlatform},
		{Permission: "shop:group", Title: "店铺运营", Path: "", Icon: "Goods", Sort: 20, Scope: model.MenuScopeTenant},
		{Permission: "product:list", ParentPerm: "shop:group", Title: "商品管理", Path: "/products", Icon: "Goods", Sort: 21, Scope: model.MenuScopeTenant},
		{Permission: "coupon:list", ParentPerm: "shop:group", Title: "优惠券管理", Path: "/coupons", Icon: "Ticket", Sort: 22, Scope: model.MenuScopeTenant},
		{Permission: "order:list", ParentPerm: "shop:group", Title: "订单管理", Path: "/orders", Icon: "List", Sort: 23, Scope: model.MenuScopeTenant},
		{Permission: "user:group", Title: "用户中心", Path: "", Icon: "User", Sort: 30, Scope: model.MenuScopeTenant},
		{Permission: "user:list", ParentPerm: "user:group", Title: "用户管理", Path: "/users", Icon: "User", Sort: 31, Scope: model.MenuScopeTenant},
		{Permission: "admin:tenant:list", ParentPerm: "user:group", Title: "账号管理", Path: "/admins", Icon: "Avatar", Sort: 32, Scope: model.MenuScopeTenant},
	}
}

func (s *RBACService) ensureMenuHierarchy(ctx context.Context) error {
	byPerm := make(map[string]uint64)
	perms := make([]string, 0, len(defaultMenuSpecs()))
	for _, spec := range defaultMenuSpecs() {
		perms = append(perms, spec.Permission)
		m, err := s.rbacRepo.GetMenuByPermission(ctx, spec.Permission)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			created, createErr := s.CreateMenu(ctx, CreateMenuReq{
				Title: spec.Title, Path: spec.Path, Icon: spec.Icon, Sort: spec.Sort,
				Scope: spec.Scope, Permission: spec.Permission, Status: 1,
			})
			if createErr != nil {
				return createErr
			}
			m = created
		} else if err != nil {
			return err
		} else {
			m.Title = spec.Title
			m.Path = spec.Path
			m.Icon = spec.Icon
			m.Sort = spec.Sort
			m.Scope = spec.Scope
			m.Status = 1
			if err := s.rbacRepo.UpdateMenu(ctx, m); err != nil {
				return err
			}
		}
		byPerm[spec.Permission] = m.ID
	}
	for _, spec := range defaultMenuSpecs() {
		parentID := uint64(0)
		if spec.ParentPerm != "" {
			parentID = byPerm[spec.ParentPerm]
		}
		if err := s.rbacRepo.ApplyMenuSpec(ctx, spec.Permission, spec.Path, spec.Scope, parentID, spec.Title, spec.Sort); err != nil {
			return err
		}
	}
	return s.rbacRepo.DisableMenusExcept(ctx, perms)
}

// --- Menu CRUD ---

type CreateMenuReq struct {
	ParentID   uint64 `json:"parent_id"`
	Title      string `json:"title" binding:"required"`
	Path       string `json:"path"`
	Icon       string `json:"icon"`
	Sort       int    `json:"sort"`
	Scope      string `json:"scope" binding:"required"`
	Permission string `json:"permission"`
	Status     int8   `json:"status"`
}

func (s *RBACService) ListMenus(ctx context.Context) ([]model.AdminMenu, error) {
	return s.rbacRepo.ListMenus(ctx)
}

func (s *RBACService) CreateMenu(ctx context.Context, req CreateMenuReq) (*model.AdminMenu, error) {
	m := &model.AdminMenu{
		ParentID: req.ParentID, Title: req.Title, Path: req.Path, Icon: req.Icon,
		Sort: req.Sort, Scope: req.Scope, Permission: req.Permission, Status: 1,
	}
	if req.Status != 0 {
		m.Status = req.Status
	}
	return m, s.rbacRepo.CreateMenu(ctx, m)
}

func (s *RBACService) UpdateMenu(ctx context.Context, id uint64, req CreateMenuReq) (*model.AdminMenu, error) {
	m, err := s.rbacRepo.GetMenu(ctx, id)
	if err != nil {
		return nil, err
	}
	m.ParentID = req.ParentID
	m.Title = req.Title
	m.Path = req.Path
	m.Icon = req.Icon
	m.Sort = req.Sort
	m.Scope = req.Scope
	m.Permission = req.Permission
	if req.Status != 0 {
		m.Status = req.Status
	}
	return m, s.rbacRepo.UpdateMenu(ctx, m)
}

func (s *RBACService) DeleteMenu(ctx context.Context, id uint64) error {
	return s.rbacRepo.DeleteMenu(ctx, id)
}

// --- Role CRUD ---

type CreateRoleReq struct {
	TenantID uint64 `json:"tenant_id"`
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Scope    string `json:"scope" binding:"required"`
}

func (s *RBACService) ListRoles(ctx context.Context, tenantID *uint64, scope string) ([]model.AdminRole, error) {
	return s.rbacRepo.ListRoles(ctx, tenantID, scope)
}

func (s *RBACService) CreateRole(ctx context.Context, req CreateRoleReq) (*model.AdminRole, error) {
	role := &model.AdminRole{
		TenantID: req.TenantID, Code: req.Code, Name: req.Name, Scope: req.Scope, Status: 1,
	}
	return role, s.rbacRepo.CreateRole(ctx, role)
}

func (s *RBACService) UpdateRole(ctx context.Context, id uint64, name string, status int8) (*model.AdminRole, error) {
	role, err := s.rbacRepo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		role.Name = name
	}
	if status != 0 {
		role.Status = status
	}
	return role, s.rbacRepo.UpdateRole(ctx, role)
}

func (s *RBACService) DeleteRole(ctx context.Context, id uint64) error {
	return s.rbacRepo.DeleteRole(ctx, id)
}

func (s *RBACService) GetRoleMenus(ctx context.Context, roleID uint64) ([]uint64, error) {
	return s.rbacRepo.GetRoleMenuIDs(ctx, roleID)
}

func (s *RBACService) SetRoleMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	return s.rbacRepo.SetRoleMenus(ctx, roleID, menuIDs)
}

// --- Admin account CRUD ---

type CreateAdminReq struct {
	TenantID uint64 `json:"tenant_id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	UserType string `json:"user_type"`
	RoleIDs  []uint64 `json:"role_ids"`
}

func (s *RBACService) ListAdmins(ctx context.Context, tenantID *uint64, page, pageSize int, keyword string) ([]model.AdminUser, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.adminRepo.List(ctx, tenantID, page, pageSize, keyword)
}

func (s *RBACService) CreateAdmin(ctx context.Context, req CreateAdminReq) (*model.AdminUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	userType := req.UserType
	if userType == "" {
		if req.TenantID == 0 {
			userType = model.AdminUserTypePlatform
		} else {
			userType = model.AdminUserTypeTenant
		}
	}
	admin := &model.AdminUser{
		TenantID: req.TenantID, Username: req.Username, Password: string(hash),
		Nickname: req.Nickname, UserType: userType, Status: 1,
	}
	if err := s.adminRepo.Create(ctx, admin); err != nil {
		return nil, err
	}
	if len(req.RoleIDs) > 0 {
		if err := s.rbacRepo.SetUserRoles(ctx, admin.ID, req.RoleIDs); err != nil {
			return nil, err
		}
	}
	return admin, nil
}

func (s *RBACService) UpdateAdmin(ctx context.Context, id uint64, nickname, password string, status int8, roleIDs []uint64) (*model.AdminUser, error) {
	admin, err := s.adminRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if nickname != "" {
		admin.Nickname = nickname
	}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		admin.Password = string(hash)
	}
	if status != 0 {
		admin.Status = status
	}
	if err := s.adminRepo.Update(ctx, admin); err != nil {
		return nil, err
	}
	if roleIDs != nil {
		if err := s.rbacRepo.SetUserRoles(ctx, id, roleIDs); err != nil {
			return nil, err
		}
	}
	return admin, nil
}

func (s *RBACService) GetAdminRoleIDs(ctx context.Context, adminID uint64) ([]uint64, error) {
	return s.rbacRepo.GetUserRoleIDs(ctx, adminID)
}

func (s *RBACService) EnsureDefaultRBAC(ctx context.Context) error {
	menuCount, err := s.rbacRepo.MenuCount(ctx)
	if err != nil {
		return err
	}
	if menuCount == 0 {
		if err := seedDefaultRBAC(ctx, s); err != nil {
			return err
		}
	}
	if err := s.ensureMenuHierarchy(ctx); err != nil {
		return err
	}
	return syncDefaultRoleMenus(ctx, s)
}

func syncDefaultRoleMenus(ctx context.Context, s *RBACService) error {
	superRole, err := s.rbacRepo.GetRoleByCode(ctx, 0, "super_admin")
	if err != nil {
		return nil
	}
	tenantRole, err := s.rbacRepo.GetRoleByCode(ctx, 0, "tenant_admin")
	if err != nil {
		return nil
	}
	superIDs, err := s.rbacRepo.AllMenuIDs(ctx)
	if err != nil {
		return err
	}
	if len(superIDs) > 0 {
		if err := s.SetRoleMenus(ctx, superRole.ID, superIDs); err != nil {
			return err
		}
	}
	tenantPerms := []string{
		"dashboard:view", "shop:group", "product:list", "coupon:list", "order:list",
		"user:group", "user:list", "admin:tenant:list",
	}
	tenantIDs, err := s.rbacRepo.MenuIDsByPermissions(ctx, tenantPerms)
	if err != nil {
		return err
	}
	if len(tenantIDs) > 0 {
		if err := s.SetRoleMenus(ctx, tenantRole.ID, tenantIDs); err != nil {
			return err
		}
	}
	return nil
}

func repairDefaultRoleMenus(ctx context.Context, s *RBACService) error {
	return syncDefaultRoleMenus(ctx, s)
}

func seedDefaultRBAC(ctx context.Context, s *RBACService) error {
	if err := s.ensureMenuHierarchy(ctx); err != nil {
		return err
	}
	superRole, err := s.CreateRole(ctx, CreateRoleReq{Code: "super_admin", Name: "总管理员", Scope: model.RoleScopePlatform})
	if err != nil {
		return err
	}
	tenantRole, err := s.CreateRole(ctx, CreateRoleReq{Code: "tenant_admin", Name: "租户管理员", Scope: model.RoleScopeTenant})
	if err != nil {
		return err
	}
	superIDs, err := s.rbacRepo.AllMenuIDs(ctx)
	if err != nil {
		return err
	}
	if err := s.SetRoleMenus(ctx, superRole.ID, superIDs); err != nil {
		return err
	}
	tenantPerms := []string{
		"dashboard:view", "shop:group", "product:list", "coupon:list", "order:list",
		"user:group", "user:list", "admin:tenant:list",
	}
	tenantIDs, err := s.rbacRepo.MenuIDsByPermissions(ctx, tenantPerms)
	if err != nil {
		return err
	}
	return s.SetRoleMenus(ctx, tenantRole.ID, tenantIDs)
}
