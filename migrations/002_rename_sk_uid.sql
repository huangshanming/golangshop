-- Fix legacy column name: GORM mapped SKUID -> sk_uid, application expects sku_id
ALTER TABLE cart_items CHANGE COLUMN sk_uid sku_id BIGINT UNSIGNED NOT NULL;
ALTER TABLE order_items CHANGE COLUMN sk_uid sku_id BIGINT UNSIGNED NOT NULL;
