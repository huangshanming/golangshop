package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	adminhandler "goshop/internal/admin/handler"
	apphandler "goshop/internal/app/handler"
	"goshop/internal/auth"
	"goshop/internal/config"
	"goshop/internal/domain/model"
	"goshop/internal/domain/service"
	"goshop/internal/infra/database"
	"goshop/internal/infra/mq"
	infraredis "goshop/internal/infra/redis"
	"goshop/internal/infra/storage"
	"goshop/internal/payment/alipay"
	"goshop/internal/payment/wechat"
	"goshop/internal/repository"
	"goshop/internal/server"
	"goshop/pkg/jwtutil"
	"goshop/pkg/logger"
)

type App struct {
	Server    *server.Server
	DB        *gorm.DB
	Redis     *redis.Client
	Publisher *mq.Publisher
	Consumer  *mq.Consumer
}

func NewApp(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	if err := logger.Init(cfg.Server.Mode); err != nil {
		return nil, err
	}

	db, err := database.NewMySQL(&cfg.Database)
	if err != nil {
		return nil, err
	}
	if err := database.AutoMigrate(db); err != nil {
		return nil, err
	}
	if err := seedDefaultTenant(db); err != nil {
		return nil, err
	}

	rdb, err := infraredis.NewClient(&cfg.Redis)
	if err != nil {
		logger.Info("redis unavailable, continuing without redis: " + err.Error())
		rdb = nil
	}

	var publisher *mq.Publisher
	var consumer *mq.Consumer
	publisher, err = mq.NewPublisher(&cfg.RabbitMQ)
	if err != nil {
		logger.Info("rabbitmq unavailable, continuing without mq: " + err.Error())
	} else {
		consumer, _ = mq.NewConsumer(&cfg.RabbitMQ)
	}

	enforcer, err := auth.NewEnforcer(db)
	if err != nil {
		return nil, fmt.Errorf("casbin: %w", err)
	}

	jwtMgr := jwtutil.NewManager(cfg.JWT.AppSecret, cfg.JWT.AdminSecret, cfg.JWT.AppExpireHours, cfg.JWT.AdminExpireHours)

	userRepo := repository.NewUserRepo(db)
	adminRepo := repository.NewAdminRepo(db)
	rbacRepo := repository.NewRBACRepo(db)
	productRepo := repository.NewProductRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	cartRepo := repository.NewCartRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	notifyRepo := repository.NewPayNotifyRepo(db)
	balanceRepo := repository.NewBalanceRepo(db)

	tenantRepo := repository.NewTenantRepo(db)
	tenantSvc := service.NewTenantService(tenantRepo)

	userSvc := service.NewUserService(userRepo, adminRepo, balanceRepo, orderRepo, jwtMgr)
	rbacSvc := service.NewRBACService(rbacRepo, adminRepo, tenantRepo, jwtMgr)
	if err := rbacSvc.EnsureDefaultRBAC(context.Background()); err != nil {
		return nil, fmt.Errorf("rbac seed: %w", err)
	}
	if err := auth.SyncPolicies(enforcer, db); err != nil {
		return nil, fmt.Errorf("casbin sync: %w", err)
	}

	addressRepo := repository.NewAddressRepo(db)
	couponRepo := repository.NewCouponRepo(db)
	freightRepo := repository.NewFreightRepo(db)
	afterSaleRepo := repository.NewAfterSaleRepo(db)

	platformRepo := repository.NewPlatformRepo(db)
	platformSvc := service.NewPlatformService(platformRepo, productRepo, tenantRepo)
	if err := platformSvc.EnsurePlatformCategories(context.Background()); err != nil {
		return nil, fmt.Errorf("platform categories: %w", err)
	}
	if err := platformSvc.EnsureDefaultSiteSettings(context.Background()); err != nil {
		return nil, fmt.Errorf("platform settings: %w", err)
	}
	if err := platformSvc.EnsureDefaultBanners(context.Background()); err != nil {
		return nil, fmt.Errorf("platform banners: %w", err)
	}
	if err := platformSvc.EnsureDefaultFeatured(context.Background()); err != nil {
		return nil, fmt.Errorf("platform featured: %w", err)
	}

	categorySvc := service.NewCategoryService(platformSvc)
	productSvc := service.NewProductService(productRepo, categoryRepo, platformRepo, tenantRepo)
	cartSvc := service.NewCartService(cartRepo, productRepo)
	addressSvc := service.NewAddressService(addressRepo)
	couponSvc := service.NewCouponService(couponRepo)
	balanceSvc := service.NewBalanceService(balanceRepo, userRepo, orderRepo)
	checkoutSvc := service.NewCheckoutService(
		orderRepo, productRepo, cartRepo, addressRepo, couponRepo, freightRepo, balanceRepo, userRepo, publisher, cfg.JWT.AppSecret,
	)
	afterSaleSvc := service.NewAfterSaleService(afterSaleRepo, orderRepo, balanceRepo, productRepo)
	orderSvc := service.NewOrderService(checkoutSvc, orderRepo, productRepo, couponRepo, balanceRepo, afterSaleRepo, publisher)
	wechatClient := wechat.NewClient(cfg.Payment.Wechat)
	alipayClient := alipay.NewClient(cfg.Payment.Alipay)
	paySvc := service.NewPaymentService(orderSvc, balanceRepo, notifyRepo, wechatClient, alipayClient, cfg.Payment)

	var uploadSvc *service.UploadService
	minioClient, err := storage.NewMinIOClient(&cfg.MinIO)
	if err != nil {
		logger.Info("minio unavailable, upload disabled: " + err.Error())
	} else {
		uploadSvc = service.NewUploadService(minioClient)
	}

	handlers := &server.Handlers{
		AppAuth:       apphandler.NewAuthHandler(userSvc, tenantSvc),
		AppUser:       apphandler.NewUserHandler(userSvc),
		AppProduct:    apphandler.NewProductHandler(productSvc),
		AppHome:       apphandler.NewHomeHandler(platformSvc, productSvc),
		AppCategory:   apphandler.NewCategoryHandler(categorySvc),
		AppCart:       apphandler.NewCartHandler(cartSvc),
		AppOrder:      apphandler.NewOrderHandler(orderSvc, paySvc),
		AppAddress:    apphandler.NewAddressHandler(addressSvc),
		AppCoupon:     apphandler.NewCouponHandler(couponSvc),
		AppAfterSale:  apphandler.NewAfterSaleHandler(afterSaleSvc),
		AppPayNotify:  apphandler.NewPayNotifyHandler(paySvc),
		AppTenant:     apphandler.NewTenantHandler(tenantSvc),
		AdminAuth:     adminhandler.NewAuthHandler(rbacSvc, enforcer, db),
		AdminRBAC:     adminhandler.NewRBACHandler(rbacSvc),
		AdminProduct:  adminhandler.NewProductHandler(productSvc, platformSvc),
		AdminPlatform: adminhandler.NewPlatformHandler(platformSvc, productSvc),
		AdminCategory: adminhandler.NewCategoryHandler(categorySvc),
		AdminUpload:   adminhandler.NewUploadHandler(uploadSvc),
		AdminUser:     adminhandler.NewAdminUserHandler(balanceSvc),
		AdminOrder:    adminhandler.NewAdminOrderHandler(orderSvc),
		AdminCoupon:   adminhandler.NewCouponHandler(couponSvc),
		AdminTenant:   adminhandler.NewAdminTenantHandler(tenantSvc),
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			_ = orderSvc.CloseExpiredOrders(context.Background())
		}
	}()

	if consumer != nil {
		_ = consumer.ConsumeOrderPaid(func(evt mq.OrderPaidEvent) error {
			logger.Info(fmt.Sprintf("order paid event: %+v", evt))
			return nil
		})
	}

	router := server.NewRouter(server.RouterDeps{
		Cfg: cfg, DB: db, JWT: jwtMgr, Enforcer: enforcer, Handlers: handlers,
	})
	srv := server.NewServer(cfg, router)

	return &App{Server: srv, DB: db, Redis: rdb, Publisher: publisher, Consumer: consumer}, nil
}

func seedDefaultTenant(db *gorm.DB) error {
	var count int64
	db.Model(&model.Tenant{}).Count(&count)
	if count > 0 {
		return nil
	}
	return db.Create(&model.Tenant{Name: "默认租户", Code: "default", Status: 1}).Error
}

func (a *App) Close() {
	logger.Sync()
	if a.Publisher != nil {
		_ = a.Publisher.Close()
	}
	if a.Consumer != nil {
		_ = a.Consumer.Close()
	}
}
