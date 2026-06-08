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
	AppProduct    *apphandler.ProductHandler
	AppCategory   *apphandler.CategoryHandler
	AppCart       *apphandler.CartHandler
	AppOrder      *apphandler.OrderHandler
	AppAddress    *apphandler.AddressHandler
	AppCoupon     *apphandler.CouponHandler
	AppAfterSale  *apphandler.AfterSaleHandler
	AppPayNotify  *apphandler.PayNotifyHandler
	AdminAuth     *adminhandler.AuthHandler
	AdminProduct  *adminhandler.ProductHandler
	AdminCategory *adminhandler.CategoryHandler
	AdminUpload   *adminhandler.UploadHandler
	AdminUser     *adminhandler.AdminUserHandler
	AdminOrder    *adminhandler.AdminOrderHandler
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
		}
	}

	admin := api.Group("/admin")
	admin.Use(middleware.Tenant())
	{
		admin.POST("/auth/login", deps.Handlers.AdminAuth.Login)

		secured := admin.Group("")
		secured.Use(middleware.AdminJWT(deps.JWT))
		if deps.Enforcer != nil {
			secured.Use(middleware.CasbinAuth(deps.Enforcer))
		}
		{
			secured.GET("/products", deps.Handlers.AdminProduct.List)
			secured.POST("/products", deps.Handlers.AdminProduct.Create)
			secured.GET("/products/:id", deps.Handlers.AdminProduct.Detail)
			secured.PUT("/products/:id", deps.Handlers.AdminProduct.Update)
			secured.PATCH("/products/:id/status", deps.Handlers.AdminProduct.UpdateStatus)
			secured.GET("/categories", deps.Handlers.AdminCategory.List)
			secured.POST("/upload", deps.Handlers.AdminUpload.Upload)

			secured.GET("/users", deps.Handlers.AdminUser.List)
			secured.POST("/users/:id/balance", deps.Handlers.AdminUser.AdjustBalance)
			secured.GET("/users/:id/balance-logs", deps.Handlers.AdminUser.BalanceLogs)

			secured.GET("/orders", deps.Handlers.AdminOrder.List)
			secured.POST("/orders/:order_no/ship", deps.Handlers.AdminOrder.Ship)
		}
	}

	return r
}
