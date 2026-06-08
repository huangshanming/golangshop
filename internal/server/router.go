package server

import (
	"net/http"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"

	adminhandler "goshop/internal/admin/handler"
	apphandler "goshop/internal/app/handler"
	"goshop/internal/config"
	"goshop/internal/infra/database"
	"goshop/internal/middleware"
	"goshop/pkg/jwtutil"
)

type Handlers struct {
	AppAuth       *apphandler.AuthHandler
	AppUser       *apphandler.UserHandler
	AppProduct    *apphandler.ProductHandler
	AppHome       *apphandler.HomeHandler
	AppCategory   *apphandler.CategoryHandler
	AppCart       *apphandler.CartHandler
	AppOrder      *apphandler.OrderHandler
	AppAddress    *apphandler.AddressHandler
	AppCoupon     *apphandler.CouponHandler
	AppAfterSale  *apphandler.AfterSaleHandler
	AppPayNotify  *apphandler.PayNotifyHandler
	AppTenant     *apphandler.TenantHandler
	AdminAuth     *adminhandler.AuthHandler
	AdminRBAC     *adminhandler.RBACHandler
	AdminProduct  *adminhandler.ProductHandler
	AdminCategory *adminhandler.CategoryHandler
	AdminUpload   *adminhandler.UploadHandler
	AdminUser     *adminhandler.AdminUserHandler
	AdminOrder    *adminhandler.AdminOrderHandler
	AdminCoupon   *adminhandler.CouponHandler
	AdminTenant   *adminhandler.AdminTenantHandler
	AdminPlatform *adminhandler.PlatformHandler
}

type RouterDeps struct {
	Cfg      *config.Config
	DB       *gorm.DB
	Redis    interface{}
	JWT      *jwtutil.Manager
	Enforcer *casbin.Enforcer
	Handlers *Handlers
}

func NewRouter(deps RouterDeps) *gin.Engine {
	if deps.Cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.TraceID())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Tenant-Id", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if err := database.Ping(deps.DB); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "db": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := r.Group("/api/v1")
	api.GET("/public/tenants", deps.Handlers.AppTenant.ListPublic)
	api.GET("/public/site-settings", deps.Handlers.AppHome.SiteSettings)
	api.GET("/public/banners", deps.Handlers.AppHome.Banners)
	api.GET("/public/categories", deps.Handlers.AppHome.Categories)
	api.GET("/public/categories/tree", deps.Handlers.AppHome.CategoryTree)
	api.GET("/public/products/featured", deps.Handlers.AppHome.FeaturedProducts)
	api.GET("/public/products", deps.Handlers.AppProduct.ListPublic)
	api.GET("/public/products/:id", deps.Handlers.AppProduct.DetailPublic)

	pay := api.Group("/pay")
	{
		pay.POST("/notify/wechat", deps.Handlers.AppPayNotify.Wechat)
		pay.POST("/notify/alipay", deps.Handlers.AppPayNotify.Alipay)
	}

	app := api.Group("/app")
	app.Use(middleware.Tenant())
	{
		app.POST("/auth/login", deps.Handlers.AppAuth.Login)

		app.GET("/products", deps.Handlers.AppProduct.List)
		app.GET("/products/:id", deps.Handlers.AppProduct.Detail)
		app.GET("/categories", deps.Handlers.AppCategory.List)

		auth := app.Group("")
		auth.Use(middleware.AppJWT(deps.JWT))
		{
			auth.GET("/cart", deps.Handlers.AppCart.List)
			auth.POST("/cart", deps.Handlers.AppCart.Add)
			auth.PATCH("/cart/:id", deps.Handlers.AppCart.Patch)
			auth.POST("/cart/check-all", deps.Handlers.AppCart.CheckAll)
			auth.DELETE("/cart/:id", deps.Handlers.AppCart.Remove)

			auth.GET("/addresses", deps.Handlers.AppAddress.List)
			auth.POST("/addresses", deps.Handlers.AppAddress.Create)
			auth.PUT("/addresses/:id", deps.Handlers.AppAddress.Update)
			auth.DELETE("/addresses/:id", deps.Handlers.AppAddress.Delete)

			auth.GET("/coupons/available", deps.Handlers.AppCoupon.Available)
			auth.GET("/coupons/mine", deps.Handlers.AppCoupon.Mine)
			auth.POST("/coupons/:id/receive", deps.Handlers.AppCoupon.Receive)

			auth.POST("/orders/preview", deps.Handlers.AppOrder.Preview)
			auth.POST("/orders", deps.Handlers.AppOrder.Create)
			auth.GET("/orders", deps.Handlers.AppOrder.List)
			auth.GET("/orders/:order_no", deps.Handlers.AppOrder.Detail)
			auth.POST("/orders/pay", deps.Handlers.AppOrder.Pay)
			auth.POST("/orders/:order_no/mock-pay", deps.Handlers.AppOrder.MockPay)
			auth.POST("/orders/:order_no/cancel", deps.Handlers.AppOrder.Cancel)
			auth.POST("/orders/:order_no/confirm-receive", deps.Handlers.AppOrder.ConfirmReceive)
			auth.POST("/orders/:order_no/review", deps.Handlers.AppOrder.Review)

			auth.POST("/after-sales", deps.Handlers.AppAfterSale.Create)

			auth.GET("/user/me", deps.Handlers.AppUser.Me)
			auth.PUT("/user/profile", deps.Handlers.AppUser.UpdateProfile)
			auth.GET("/user/balance-logs", deps.Handlers.AppUser.BalanceLogs)
		}
	}

	admin := api.Group("/admin")
	{
		admin.POST("/auth/login", deps.Handlers.AdminAuth.Login)

		secured := admin.Group("")
		secured.Use(middleware.AdminJWT(deps.JWT))
		if deps.Enforcer != nil {
			secured.Use(middleware.CasbinAuth(deps.Enforcer))
		}
		{
			secured.GET("/auth/me", deps.Handlers.AdminAuth.Me)
			secured.POST("/auth/enter-tenant", deps.Handlers.AdminAuth.EnterTenant)
			secured.POST("/auth/exit-tenant", deps.Handlers.AdminAuth.ExitTenant)

			secured.GET("/menus", deps.Handlers.AdminRBAC.ListMenus)
			secured.POST("/menus", deps.Handlers.AdminRBAC.CreateMenu)
			secured.PUT("/menus/:id", deps.Handlers.AdminRBAC.UpdateMenu)
			secured.DELETE("/menus/:id", deps.Handlers.AdminRBAC.DeleteMenu)

			secured.GET("/roles", deps.Handlers.AdminRBAC.ListRoles)
			secured.POST("/roles", deps.Handlers.AdminRBAC.CreateRole)
			secured.PUT("/roles/:id", deps.Handlers.AdminRBAC.UpdateRole)
			secured.DELETE("/roles/:id", deps.Handlers.AdminRBAC.DeleteRole)
			secured.GET("/roles/:id/menus", deps.Handlers.AdminRBAC.GetRoleMenus)
			secured.PUT("/roles/:id/menus", deps.Handlers.AdminRBAC.SetRoleMenus)

			secured.GET("/admins", deps.Handlers.AdminRBAC.ListAdmins)
			secured.POST("/admins", deps.Handlers.AdminRBAC.CreateAdmin)
			secured.PUT("/admins/:id", deps.Handlers.AdminRBAC.UpdateAdmin)
			secured.GET("/admins/:id/roles", deps.Handlers.AdminRBAC.GetAdminRoles)

			secured.GET("/tenants", deps.Handlers.AdminTenant.List)
			secured.POST("/tenants", deps.Handlers.AdminTenant.Create)
			secured.GET("/tenants/:id", deps.Handlers.AdminTenant.Detail)
			secured.GET("/tenants/current", deps.Handlers.AdminTenant.Current)

			secured.GET("/platform-categories", deps.Handlers.AdminPlatform.ListCategories)
			secured.POST("/platform-categories", deps.Handlers.AdminPlatform.CreateCategory)
			secured.PUT("/platform-categories/:id", deps.Handlers.AdminPlatform.UpdateCategory)
			secured.DELETE("/platform-categories/:id", deps.Handlers.AdminPlatform.DeleteCategory)
			secured.GET("/banners", deps.Handlers.AdminPlatform.ListBanners)
			secured.POST("/banners", deps.Handlers.AdminPlatform.CreateBanner)
			secured.PUT("/banners/:id", deps.Handlers.AdminPlatform.UpdateBanner)
			secured.DELETE("/banners/:id", deps.Handlers.AdminPlatform.DeleteBanner)
			secured.GET("/site-settings", deps.Handlers.AdminPlatform.GetSiteSettings)
			secured.PUT("/site-settings", deps.Handlers.AdminPlatform.UpdateSiteSettings)
			secured.GET("/home-featured", deps.Handlers.AdminPlatform.ListFeatured)
			secured.POST("/home-featured", deps.Handlers.AdminPlatform.AddFeatured)
			secured.DELETE("/home-featured/:id", deps.Handlers.AdminPlatform.DeleteFeatured)
			secured.PUT("/home-featured/sort", deps.Handlers.AdminPlatform.UpdateFeaturedSort)
			secured.GET("/products/search", deps.Handlers.AdminPlatform.SearchProducts)

			biz := secured.Group("")
			biz.Use(middleware.AdminRequireTenant())
			{
				biz.GET("/products", deps.Handlers.AdminProduct.List)
				biz.POST("/products", deps.Handlers.AdminProduct.Create)
				biz.GET("/products/:id", deps.Handlers.AdminProduct.Detail)
				biz.PUT("/products/:id", deps.Handlers.AdminProduct.Update)
				biz.PATCH("/products/:id/status", deps.Handlers.AdminProduct.UpdateStatus)
				biz.GET("/categories", deps.Handlers.AdminCategory.List)
				biz.POST("/upload", deps.Handlers.AdminUpload.Upload)

				biz.GET("/users", deps.Handlers.AdminUser.List)
				biz.POST("/users/:id/balance", deps.Handlers.AdminUser.AdjustBalance)
				biz.GET("/users/:id/balance-logs", deps.Handlers.AdminUser.BalanceLogs)

				biz.GET("/orders", deps.Handlers.AdminOrder.List)
				biz.POST("/orders/:order_no/ship", deps.Handlers.AdminOrder.Ship)

				biz.GET("/coupons", deps.Handlers.AdminCoupon.List)
				biz.POST("/coupons", deps.Handlers.AdminCoupon.Create)
				biz.GET("/coupons/:id", deps.Handlers.AdminCoupon.Detail)
				biz.PUT("/coupons/:id", deps.Handlers.AdminCoupon.Update)
				biz.PATCH("/coupons/:id/status", deps.Handlers.AdminCoupon.UpdateStatus)
			}
		}
	}

	return r
}
