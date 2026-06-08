package model

const (
	CouponTypeFixed   = "fixed"
	CouponTypePercent = "percent"

	UserCouponUnused  = "unused"
	UserCouponUsed    = "used"
	UserCouponExpired = "expired"
	UserCouponInvalid = "invalid"
)

type Coupon struct {
	BaseModel
	TenantScoped
	Name            string `gorm:"size:128;not null" json:"name"`
	Type            string `gorm:"size:32;not null" json:"type"`
	ThresholdAmount int64  `gorm:"default:0" json:"threshold_amount"`
	DiscountValue   int64  `gorm:"not null" json:"discount_value"`
	MaxDiscount     int64  `gorm:"default:0" json:"max_discount"`
	ValidFrom       int64  `json:"valid_from"`
	ValidTo         int64  `json:"valid_to"`
	TotalCount      int64  `gorm:"default:0" json:"total_count"`
	ReceivedCount   int64  `gorm:"default:0" json:"received_count"`
	PerUserLimit    int    `gorm:"default:1" json:"per_user_limit"`
	Status          int8   `gorm:"default:1" json:"status"`
}

func (Coupon) TableName() string { return "coupons" }

type UserCoupon struct {
	BaseModel
	TenantScoped
	UserID      uint64 `gorm:"index;not null" json:"user_id"`
	CouponID    uint64 `gorm:"index;not null" json:"coupon_id"`
	Status      string `gorm:"size:32;default:unused" json:"status"`
	UsedOrderNo string `gorm:"size:64" json:"used_order_no"`
	ReceivedAt  int64  `json:"received_at"`
	UsedAt      *int64 `json:"used_at"`
}

func (UserCoupon) TableName() string { return "user_coupons" }

type OrderPromotion struct {
	BaseModel
	TenantScoped
	OrderID        uint64 `gorm:"index;not null" json:"order_id"`
	UserCouponID   uint64 `gorm:"index" json:"user_coupon_id"`
	CouponName     string `gorm:"size:128" json:"coupon_name"`
	DiscountAmount int64  `gorm:"not null" json:"discount_amount"`
}

func (OrderPromotion) TableName() string { return "order_promotions" }
