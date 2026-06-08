package model

const (
	OrderStatusPending   = 0 // 待付款
	OrderStatusPaid      = 1 // 待发货
	OrderStatusShipped   = 2 // 待收货
	OrderStatusCompleted = 3 // 已完成
	OrderStatusCancelled = 4 // 已取消
	OrderStatusClosed    = 5 // 已关闭（超时/售后完结）

	ReviewStatusNone    = 0
	ReviewStatusPending = 1
	ReviewStatusDone    = 2
)

type Order struct {
	BaseModel
	TenantScoped
	OrderNo         string `gorm:"size:64;uniqueIndex;not null" json:"order_no"`
	ClientOrderNo   string `gorm:"size:64;uniqueIndex:idx_tenant_client_order,priority:2" json:"client_order_no"`
	UserID          uint64 `gorm:"index;not null" json:"user_id"`
	Status          int8   `gorm:"default:0" json:"status"`
	GoodsAmount     int64  `gorm:"not null" json:"goods_amount"`
	DiscountAmount  int64  `gorm:"default:0" json:"discount_amount"`
	FreightAmount   int64  `gorm:"default:0" json:"freight_amount"`
	TotalAmount     int64  `gorm:"not null" json:"total_amount"` // 兼容：等于应付 pay_amount
	PayAmount       int64  `gorm:"default:0" json:"pay_amount"`
	PayChannel      string `gorm:"size:32" json:"pay_channel"`
	PayNo           string `gorm:"size:128" json:"pay_no"`
	PaidAt          *int64 `json:"paid_at"`
	UserCouponID    uint64 `gorm:"default:0" json:"user_coupon_id"`
	AddressSnapshot string `gorm:"type:text" json:"address_snapshot"`
	ExpireAt        int64  `gorm:"default:0" json:"expire_at"`
	ShippedAt       *int64 `json:"shipped_at"`
	ReceivedAt      *int64 `json:"received_at"`
	CompletedAt     *int64 `json:"completed_at"`
	ReviewStatus    int8   `gorm:"default:0" json:"review_status"`
	CancelReason    string `gorm:"size:256" json:"cancel_reason"`
	CancelledAt     *int64 `json:"cancelled_at"`
	Remark          string `gorm:"size:512" json:"remark"`
	Items           []OrderItem      `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Promotions      []OrderPromotion `gorm:"foreignKey:OrderID" json:"promotions,omitempty"`
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	BaseModel
	TenantScoped
	OrderID       uint64 `gorm:"index;not null" json:"order_id"`
	ProductID     uint64 `gorm:"not null" json:"product_id"`
	SkuID         uint64 `gorm:"column:sku_id;not null" json:"sku_id"`
	ProductName   string `gorm:"size:256" json:"product_name"`
	SKUName       string `gorm:"size:256" json:"sku_name"`
	CoverImage    string `gorm:"size:512" json:"cover_image"`
	SpecValues    string `gorm:"type:text" json:"spec_values"`
	Price         int64  `gorm:"not null" json:"price"`
	OriginalPrice int64  `gorm:"default:0" json:"original_price"`
	LineAmount    int64  `gorm:"not null" json:"line_amount"`
	Quantity      int    `gorm:"not null" json:"quantity"`
}

func (OrderItem) TableName() string { return "order_items" }

type PayNotifyLog struct {
	BaseModel
	TenantID  uint64 `gorm:"index" json:"tenant_id"`
	Channel   string `gorm:"size:32;index" json:"channel"`
	NotifyID  string `gorm:"size:128;uniqueIndex" json:"notify_id"`
	OrderNo   string `gorm:"size:64;index" json:"order_no"`
	Payload   string `gorm:"type:text" json:"payload"`
	Processed bool   `gorm:"default:false" json:"processed"`
}

func (PayNotifyLog) TableName() string { return "pay_notify_logs" }
