package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"goshop/internal/domain/model"
	"goshop/internal/infra/mq"
	"goshop/internal/repository"
)

type OrderService struct {
	checkout    *CheckoutService
	orderRepo   *repository.OrderRepo
	productRepo *repository.ProductRepo
	couponRepo  *repository.CouponRepo
	balanceRepo *repository.BalanceRepo
	afterSaleRepo *repository.AfterSaleRepo
	publisher   *mq.Publisher
}

func NewOrderService(
	checkout *CheckoutService,
	orderRepo *repository.OrderRepo,
	productRepo *repository.ProductRepo,
	couponRepo *repository.CouponRepo,
	balanceRepo *repository.BalanceRepo,
	afterSaleRepo *repository.AfterSaleRepo,
	publisher *mq.Publisher,
) *OrderService {
	return &OrderService{
		checkout: checkout, orderRepo: orderRepo, productRepo: productRepo,
		couponRepo: couponRepo, balanceRepo: balanceRepo, afterSaleRepo: afterSaleRepo, publisher: publisher,
	}
}

func (s *OrderService) Preview(ctx context.Context, tenantID, userID uint64, req PreviewReq) (*PreviewResult, error) {
	return s.checkout.Preview(ctx, tenantID, userID, req)
}

func (s *OrderService) Create(ctx context.Context, tenantID, userID uint64, req CreateOrderReq) (*model.Order, error) {
	return s.checkout.Create(ctx, tenantID, userID, req)
}

func (s *OrderService) List(ctx context.Context, tenantID, userID uint64, page, pageSize int, tab string) ([]OrderVO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var status *int8
	reviewPending := false
	switch tab {
	case "pending_pay":
		v := int8(model.OrderStatusPending)
		status = &v
	case "paid":
		v := int8(model.OrderStatusPaid)
		status = &v
	case "shipped":
		v := int8(model.OrderStatusShipped)
		status = &v
	case "review":
		reviewPending = true
	}
	list, total, err := s.orderRepo.ListByUser(ctx, tenantID, userID, page, pageSize, status, reviewPending)
	if err != nil {
		return nil, 0, err
	}
	vos := make([]OrderVO, 0, len(list))
	for _, o := range list {
		vos = append(vos, s.toOrderVO(ctx, tenantID, &o))
	}
	return vos, total, nil
}

func (s *OrderService) Get(ctx context.Context, tenantID, userID uint64, orderNo string) (*OrderVO, error) {
	o, err := s.orderRepo.FindByOrderNoForUser(ctx, tenantID, userID, orderNo)
	if err != nil {
		return nil, err
	}
	vo := s.toOrderVO(ctx, tenantID, o)
	return &vo, nil
}

func (s *OrderService) GetInternal(ctx context.Context, tenantID uint64, orderNo string) (*model.Order, error) {
	return s.orderRepo.FindByOrderNo(ctx, tenantID, orderNo)
}

func (s *OrderService) toOrderVO(ctx context.Context, tenantID uint64, o *model.Order) OrderVO {
	hasAfterSale := false
	afterSaleType := ""
	if s.afterSaleRepo != nil {
		active, _ := s.afterSaleRepo.HasActiveByOrder(ctx, tenantID, o.ID)
		hasAfterSale = active
	}
	return OrderVO{
		Order:         *o,
		DisplayStatus: formatDisplayStatus(o.Status, o.ReviewStatus, hasAfterSale, afterSaleType),
	}
}

func (s *OrderService) MarkPaid(ctx context.Context, tenantID uint64, orderNo, payNo, channel string, amount int64) error {
	paidAt := time.Now().Unix()
	if err := s.orderRepo.UpdatePaid(ctx, tenantID, orderNo, payNo, channel, amount, paidAt); err != nil {
		return err
	}
	if s.publisher != nil {
		return s.publisher.PublishOrderPaid(ctx, mq.OrderPaidEvent{
			TenantID: tenantID, OrderNo: orderNo, PayNo: payNo, Channel: channel, Amount: amount,
		})
	}
	return nil
}

func (s *OrderService) Cancel(ctx context.Context, tenantID, userID uint64, orderNo, reason string) error {
	o, err := s.orderRepo.FindByOrderNoForUser(ctx, tenantID, userID, orderNo)
	if err != nil {
		return err
	}
	if o.Status != model.OrderStatusPending {
		return fmt.Errorf("order cannot be cancelled")
	}
	now := time.Now().Unix()
	return s.orderRepo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.orderRepo.CancelPending(ctx, tenantID, userID, orderNo, reason, now); err != nil {
			return err
		}
		for _, item := range o.Items {
			if err := s.productRepo.RestoreStockWithTx(tx, tenantID, item.SkuID, item.Quantity); err != nil {
				return err
			}
		}
		if o.UserCouponID > 0 {
			return s.couponRepo.RestoreUserCoupon(ctx, tenantID, o.UserCouponID)
		}
		return nil
	})
}

func (s *OrderService) ConfirmReceive(ctx context.Context, tenantID, userID uint64, orderNo string) error {
	return s.orderRepo.ConfirmReceive(ctx, tenantID, userID, orderNo, time.Now().Unix())
}

func (s *OrderService) Review(ctx context.Context, tenantID, userID uint64, orderNo string) error {
	return s.orderRepo.MarkReviewed(ctx, tenantID, userID, orderNo)
}

func (s *OrderService) CloseExpiredOrders(ctx context.Context) error {
	now := time.Now().Unix()
	list, err := s.orderRepo.ListExpiredPending(ctx, now)
	if err != nil {
		return err
	}
	for _, o := range list {
		_ = s.closeOrder(ctx, &o, now)
	}
	return nil
}

func (s *OrderService) closeOrder(ctx context.Context, o *model.Order, now int64) error {
	return s.orderRepo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.orderRepo.CloseExpired(ctx, o.ID, now); err != nil {
			return err
		}
		for _, item := range o.Items {
			if err := s.productRepo.RestoreStockWithTx(tx, o.TenantID, item.SkuID, item.Quantity); err != nil {
				return err
			}
		}
		if o.UserCouponID > 0 {
			return s.couponRepo.RestoreUserCoupon(ctx, o.TenantID, o.UserCouponID)
		}
		return nil
	})
}

type ShipOrderReq struct {
	ExpressCompany string `json:"express_company" binding:"required"`
	ExpressNo      string `json:"express_no" binding:"required"`
	ShipRemark     string `json:"ship_remark"`
}

func (s *OrderService) Ship(ctx context.Context, tenantID uint64, orderNo string, req ShipOrderReq) error {
	company := strings.TrimSpace(req.ExpressCompany)
	expressNo := strings.TrimSpace(req.ExpressNo)
	remark := strings.TrimSpace(req.ShipRemark)
	if company == "" {
		return fmt.Errorf("请选择物流公司")
	}
	if _, ok := model.ExpressCompanyNames[company]; !ok {
		return fmt.Errorf("物流公司无效")
	}
	if expressNo == "" {
		return fmt.Errorf("请填写物流单号")
	}
	if len(expressNo) < 6 || len(expressNo) > 64 {
		return fmt.Errorf("物流单号长度应为 6~64 位")
	}
	if utf8.RuneCountInString(remark) > 200 {
		return fmt.Errorf("发货备注过长")
	}
	o, err := s.orderRepo.FindByOrderNo(ctx, tenantID, orderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("订单不存在")
		}
		return err
	}
	if o.Status != model.OrderStatusPaid {
		return fmt.Errorf("仅待发货订单可发货")
	}
	return s.orderRepo.Ship(ctx, tenantID, orderNo, company, expressNo, remark, time.Now().Unix())
}

func (s *OrderService) ListAdmin(ctx context.Context, tenantID uint64, page, pageSize int, status *int8) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return s.orderRepo.ListAdmin(ctx, tenantID, page, pageSize, status)
}

func ParseAddressSnapshot(raw string) map[string]string {
	out := map[string]string{}
	if raw == "" {
		return out
	}
	var a model.UserAddress
	if json.Unmarshal([]byte(raw), &a) == nil {
		out["name"] = a.Name
		out["phone"] = a.Phone
		out["full"] = fmt.Sprintf("%s%s%s%s", a.Province, a.City, a.District, a.Detail)
	}
	return out
}

type AfterSaleService struct {
	afterSaleRepo *repository.AfterSaleRepo
	orderRepo     *repository.OrderRepo
	balanceRepo   *repository.BalanceRepo
	productRepo   *repository.ProductRepo
}

func NewAfterSaleService(afterSaleRepo *repository.AfterSaleRepo, orderRepo *repository.OrderRepo, balanceRepo *repository.BalanceRepo, productRepo *repository.ProductRepo) *AfterSaleService {
	return &AfterSaleService{afterSaleRepo: afterSaleRepo, orderRepo: orderRepo, balanceRepo: balanceRepo, productRepo: productRepo}
}

type CreateAfterSaleReq struct {
	OrderNo      string `json:"order_no" binding:"required"`
	OrderItemID  uint64 `json:"order_item_id"`
	Type         string `json:"type" binding:"required,oneof=exchange return_refund refund_only"`
	Reason       string `json:"reason" binding:"required"`
	RefundAmount int64  `json:"refund_amount"`
}

func (s *AfterSaleService) Create(ctx context.Context, tenantID, userID uint64, req CreateAfterSaleReq) (*model.AfterSaleOrder, error) {
	o, err := s.orderRepo.FindByOrderNoForUser(ctx, tenantID, userID, req.OrderNo)
	if err != nil {
		return nil, err
	}
	if o.Status != model.OrderStatusPaid && o.Status != model.OrderStatusShipped && o.Status != model.OrderStatusCompleted {
		return nil, fmt.Errorf("order status not allowed for after sale")
	}
	active, _ := s.afterSaleRepo.HasActiveByOrder(ctx, tenantID, o.ID)
	if active {
		return nil, fmt.Errorf("after sale already exists")
	}
	a := &model.AfterSaleOrder{
		TenantScoped: model.TenantScoped{TenantID: tenantID},
		AfterSaleNo:  "AS" + uuid.New().String()[:8],
		OrderID:      o.ID,
		OrderItemID:  req.OrderItemID,
		UserID:       userID,
		Type:         req.Type,
		Status:       model.AfterSaleStatusPending,
		Reason:       req.Reason,
		RefundAmount: req.RefundAmount,
	}
	return a, s.afterSaleRepo.Create(ctx, a)
}
