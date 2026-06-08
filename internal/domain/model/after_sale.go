package model

const (
	AfterSaleTypeExchange      = "exchange"
	AfterSaleTypeReturnRefund  = "return_refund"
	AfterSaleTypeRefundOnly    = "refund_only"

	AfterSaleStatusPending    = "pending"
	AfterSaleStatusApproved   = "approved"
	AfterSaleStatusRejected   = "rejected"
	AfterSaleStatusProcessing = "processing"
	AfterSaleStatusCompleted  = "completed"
	AfterSaleStatusClosed     = "closed"
)

type AfterSaleOrder struct {
	BaseModel
	TenantScoped
	AfterSaleNo     string `gorm:"size:64;uniqueIndex;not null" json:"after_sale_no"`
	OrderID         uint64 `gorm:"index;not null" json:"order_id"`
	OrderItemID     uint64 `gorm:"index" json:"order_item_id"`
	UserID          uint64 `gorm:"index;not null" json:"user_id"`
	Type            string `gorm:"size:32;not null" json:"type"`
	Status          string `gorm:"size:32;default:pending" json:"status"`
	Reason          string `gorm:"size:512" json:"reason"`
	EvidenceImages  string `gorm:"type:text" json:"evidence_images"`
	RefundAmount    int64  `gorm:"default:0" json:"refund_amount"`
	ExpressNo       string `gorm:"size:64" json:"express_no"`
}

func (AfterSaleOrder) TableName() string { return "after_sale_orders" }
