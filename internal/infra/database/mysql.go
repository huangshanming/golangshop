package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"goshop/internal/config"
	domainmodel "goshop/internal/domain/model"
)

func NewMySQL(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&domainmodel.Tenant{},
		&domainmodel.User{},
		&domainmodel.AdminUser{},
		&domainmodel.ProductCategory{},
		&domainmodel.Product{},
		&domainmodel.ProductSKU{},
		&domainmodel.CartItem{},
		&domainmodel.Order{},
		&domainmodel.OrderItem{},
		&domainmodel.OrderPromotion{},
		&domainmodel.PayNotifyLog{},
		&domainmodel.UserAddress{},
		&domainmodel.Coupon{},
		&domainmodel.UserCoupon{},
		&domainmodel.UserBalanceLog{},
		&domainmodel.FreightRule{},
		&domainmodel.AfterSaleOrder{},
	); err != nil {
		return err
	}
	return fixLegacySkuIDColumns(db)
}

// GORM 曾将 SKUID 映射为 sk_uid，与代码/接口约定的 sku_id 不一致。
func fixLegacySkuIDColumns(db *gorm.DB) error {
	migrator := db.Migrator()
	for _, table := range []string{"cart_items", "order_items"} {
		if migrator.HasColumn(table, "sk_uid") && !migrator.HasColumn(table, "sku_id") {
			if err := db.Exec(fmt.Sprintf(
				"ALTER TABLE `%s` CHANGE COLUMN `sk_uid` `sku_id` BIGINT UNSIGNED NOT NULL",
				table,
			)).Error; err != nil {
				return fmt.Errorf("rename %s.sk_uid: %w", table, err)
			}
		}
	}
	return nil
}

func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
