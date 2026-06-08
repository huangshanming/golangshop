package model

const (
	BalanceBizAdminAdjust = "admin_adjust"
	BalanceBizOrderPay    = "order_pay"
	BalanceBizOrderRefund = "order_refund"
	BalanceBizRecharge    = "recharge"
)

type UserBalanceLog struct {
	BaseModel
	TenantScoped
	UserID       uint64 `gorm:"index;not null" json:"user_id"`
	ChangeAmount int64  `gorm:"not null" json:"change_amount"`
	BalanceAfter int64  `gorm:"not null" json:"balance_after"`
	BizType      string `gorm:"size:32;not null" json:"biz_type"`
	BizNo        string `gorm:"size:64;index" json:"biz_no"`
	Remark       string `gorm:"size:256" json:"remark"`
	OperatorID   uint64 `gorm:"default:0" json:"operator_id"`
}

func (UserBalanceLog) TableName() string { return "user_balance_logs" }
