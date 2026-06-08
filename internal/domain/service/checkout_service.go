package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"goshop/internal/domain/model"
	"goshop/internal/infra/mq"
	"goshop/internal/repository"
)

type CheckoutService struct {
	orderRepo     *repository.OrderRepo
	productRepo   *repository.ProductRepo
	cartRepo      *repository.CartRepo
	addressRepo   *repository.AddressRepo
	couponRepo    *repository.CouponRepo
	freightRepo   *repository.FreightRepo
	balanceRepo   *repository.BalanceRepo
	userRepo      *repository.UserRepo
	publisher     *mq.Publisher
	previewSecret string
}

func NewCheckoutService(
	orderRepo *repository.OrderRepo,
	productRepo *repository.ProductRepo,
	cartRepo *repository.CartRepo,
	addressRepo *repository.AddressRepo,
	couponRepo *repository.CouponRepo,
	freightRepo *repository.FreightRepo,
	balanceRepo *repository.BalanceRepo,
	userRepo *repository.UserRepo,
	publisher *mq.Publisher,
	previewSecret string,
) *CheckoutService {
	return &CheckoutService{
		orderRepo: orderRepo, productRepo: productRepo, cartRepo: cartRepo,
		addressRepo: addressRepo, couponRepo: couponRepo, freightRepo: freightRepo,
		balanceRepo: balanceRepo, userRepo: userRepo, publisher: publisher, previewSecret: previewSecret,
	}
}

type CheckoutItemReq struct {
	SkuID    uint64 `json:"sku_id"`
	Quantity int    `json:"quantity"`
}

type PreviewReq struct {
	Items        []CheckoutItemReq `json:"items"`
	AddressID    uint64            `json:"address_id"`
	UserCouponID uint64            `json:"user_coupon_id"`
	CartItemIDs  []uint64          `json:"cart_item_ids"`
}

type PreviewLineItem struct {
	SkuID         uint64 `json:"sku_id"`
	ProductID     uint64 `json:"product_id"`
	ProductName   string `json:"product_name"`
	Spec          string `json:"spec"`
	SpecValues    string `json:"spec_values"`
	CoverImage    string `json:"cover_image"`
	Price         int64  `json:"price"`
	OriginalPrice int64  `json:"original_price"`
	Quantity      int    `json:"quantity"`
	LineAmount    int64  `json:"line_amount"`
	Stock         int64  `json:"stock"`
	StockEnough   bool   `json:"stock_enough"`
}

type PreviewAddress struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Detail   string `json:"detail"`
	FullText string `json:"full_text"`
}

type PreviewResult struct {
	Items          []PreviewLineItem `json:"items"`
	Address        *PreviewAddress   `json:"address"`
	GoodsAmount    int64             `json:"goods_amount"`
	DiscountAmount int64             `json:"discount_amount"`
	FreightAmount  int64             `json:"freight_amount"`
	PayAmount      int64             `json:"pay_amount"`
	UserCouponID   uint64            `json:"user_coupon_id"`
	CouponName     string            `json:"coupon_name"`
	PreviewToken   string            `json:"preview_token"`
	Balance        int64             `json:"balance"`
}

type CreateOrderReq struct {
	ClientOrderNo string `json:"client_order_no" binding:"required"`
	AddressID     uint64 `json:"address_id" binding:"required"`
	UserCouponID  uint64 `json:"user_coupon_id"`
	PreviewToken  string `json:"preview_token" binding:"required"`
	PayChannel    string `json:"pay_channel" binding:"required,oneof=balance wechat alipay"`
	Remark        string `json:"remark"`
	CartItemIDs   []uint64 `json:"cart_item_ids"`
	Items         []CheckoutItemReq `json:"items" binding:"required,min=1"`
}

type OrderVO struct {
	model.Order
	DisplayStatus string `json:"display_status"`
}

func (s *CheckoutService) Preview(ctx context.Context, tenantID, userID uint64, req PreviewReq) (*PreviewResult, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("items required")
	}
	lines, goodsAmount, signItems, err := s.buildLines(ctx, tenantID, req.Items)
	if err != nil {
		return nil, err
	}
	var addr *PreviewAddress
	if req.AddressID > 0 {
		a, err := s.addressRepo.GetByID(ctx, tenantID, userID, req.AddressID)
		if err != nil {
			return nil, fmt.Errorf("address invalid")
		}
		addr = toPreviewAddress(a)
	} else {
		a, err := s.addressRepo.GetDefault(ctx, tenantID, userID)
		if err == nil {
			addr = toPreviewAddress(a)
		}
	}
	discount, couponName, err := s.calcDiscount(ctx, tenantID, userID, req.UserCouponID, goodsAmount)
	if err != nil {
		return nil, err
	}
	freight, err := s.calcFreight(ctx, tenantID, goodsAmount)
	if err != nil {
		return nil, err
	}
	payAmount := goodsAmount - discount + freight
	effectiveAddressID := req.AddressID
	if effectiveAddressID == 0 && addr != nil {
		effectiveAddressID = addr.ID
	}
	exp := time.Now().Add(previewTokenTTL).Unix()
	token, err := signPreview(s.previewSecret, previewPayload{
		TenantID: tenantID, UserID: userID, AddressID: effectiveAddressID,
		UserCouponID: req.UserCouponID, Items: signItems,
		GoodsAmount: goodsAmount, DiscountAmt: discount, FreightAmt: freight,
		PayAmount: payAmount, ExpiresAt: exp,
	})
	if err != nil {
		return nil, err
	}
	balance := int64(0)
	if user, err := s.userRepo.FindByID(ctx, tenantID, userID); err == nil {
		balance = user.Balance
	}
	return &PreviewResult{
		Items: lines, Address: addr,
		GoodsAmount: goodsAmount, DiscountAmount: discount, FreightAmount: freight, PayAmount: payAmount,
		UserCouponID: req.UserCouponID, CouponName: couponName, PreviewToken: token, Balance: balance,
	}, nil
}

func (s *CheckoutService) buildLines(ctx context.Context, tenantID uint64, items []CheckoutItemReq) ([]PreviewLineItem, int64, []previewItemSign, error) {
	var goods int64
	lines := make([]PreviewLineItem, 0, len(items))
	signItems := make([]previewItemSign, 0, len(items))
	for _, it := range items {
		if it.Quantity < 1 {
			return nil, 0, nil, errors.New("invalid quantity")
		}
		sku, err := s.productRepo.GetSKU(ctx, tenantID, it.SkuID)
		if err != nil {
			return nil, 0, nil, fmt.Errorf("sku %d not found", it.SkuID)
		}
		p, _ := s.productRepo.GetByID(ctx, tenantID, sku.ProductID)
		name, cover := "", ""
		if p != nil {
			name = p.Name
			cover = p.CoverImage
		}
		lineAmount := sku.Price * int64(it.Quantity)
		goods += lineAmount
		lines = append(lines, PreviewLineItem{
			SkuID: it.SkuID, ProductID: sku.ProductID, ProductName: name,
			Spec: sku.Spec, SpecValues: sku.SpecValues, CoverImage: cover,
			Price: sku.Price, OriginalPrice: sku.OriginalPrice, Quantity: it.Quantity,
			LineAmount: lineAmount, Stock: sku.Stock, StockEnough: sku.Stock >= int64(it.Quantity),
		})
		signItems = append(signItems, previewItemSign{SkuID: it.SkuID, Quantity: it.Quantity, Price: sku.Price})
	}
	return lines, goods, signItems, nil
}

func (s *CheckoutService) calcDiscount(ctx context.Context, tenantID, userID, userCouponID uint64, goodsAmount int64) (int64, string, error) {
	if userCouponID == 0 {
		return 0, "", nil
	}
	uc, err := s.couponRepo.GetUserCoupon(ctx, tenantID, userID, userCouponID)
	if err != nil {
		return 0, "", fmt.Errorf("coupon invalid")
	}
	if uc.Status != model.UserCouponUnused {
		return 0, "", fmt.Errorf("coupon unavailable")
	}
	coupon, err := s.couponRepo.GetByID(ctx, tenantID, uc.CouponID)
	if err != nil {
		return 0, "", fmt.Errorf("coupon invalid")
	}
	now := time.Now().Unix()
	if coupon.Status != 1 || now < coupon.ValidFrom || now > coupon.ValidTo {
		return 0, "", fmt.Errorf("coupon expired")
	}
	if goodsAmount < coupon.ThresholdAmount {
		return 0, "", fmt.Errorf("coupon threshold not met")
	}
	discount := calcCouponDiscount(coupon, goodsAmount)
	return discount, coupon.Name, nil
}

func calcCouponDiscount(coupon *model.Coupon, goodsAmount int64) int64 {
	switch coupon.Type {
	case model.CouponTypeFixed:
		if coupon.DiscountValue > goodsAmount {
			return goodsAmount
		}
		return coupon.DiscountValue
	case model.CouponTypePercent:
		d := goodsAmount * coupon.DiscountValue / 100
		if coupon.MaxDiscount > 0 && d > coupon.MaxDiscount {
			return coupon.MaxDiscount
		}
		return d
	default:
		return 0
	}
}

func (s *CheckoutService) calcFreight(ctx context.Context, tenantID uint64, goodsAmount int64) (int64, error) {
	rule, err := s.freightRepo.GetDefault(ctx, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 800, nil
		}
		return 0, err
	}
	// 免邮门槛按券前商品金额判断，避免使用优惠券后反而产生运费
	if goodsAmount >= rule.FreeThreshold {
		return 0, nil
	}
	return rule.BaseFreight, nil
}

func (s *CheckoutService) Create(ctx context.Context, tenantID, userID uint64, req CreateOrderReq) (*model.Order, error) {
	if existing, err := s.orderRepo.FindByClientOrderNo(ctx, tenantID, req.ClientOrderNo); err == nil && existing != nil {
		return existing, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	payload, err := verifyPreview(s.previewSecret, req.PreviewToken, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if payload.AddressID != req.AddressID || payload.UserCouponID != req.UserCouponID {
		return nil, errors.New("preview params mismatch")
	}
	_, goodsAmount, signItems, err := s.buildLines(ctx, tenantID, req.Items)
	if err != nil {
		return nil, err
	}
	if itemsSignKey(signItems) != itemsSignKey(payload.Items) {
		return nil, errors.New("items changed, please preview again")
	}
	discount, couponName, err := s.calcDiscount(ctx, tenantID, userID, req.UserCouponID, goodsAmount)
	if err != nil {
		return nil, err
	}
	freight, err := s.calcFreight(ctx, tenantID, goodsAmount)
	if err != nil {
		return nil, err
	}
	payAmount := goodsAmount - discount + freight
	if payAmount != payload.PayAmount || goodsAmount != payload.GoodsAmount {
		return nil, errors.New("amount changed, please preview again")
	}
	addr, err := s.addressRepo.GetByID(ctx, tenantID, userID, req.AddressID)
	if err != nil {
		return nil, fmt.Errorf("address required")
	}
	addrJSON, _ := json.Marshal(addr)
	now := time.Now().Unix()
	expireAt := now + parseOrderExpireMinutes()*60

	var order *model.Order
	err = s.orderRepo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		type deductRecord struct {
			skuID uint64
			qty   int
		}
		deducted := make([]deductRecord, 0, len(req.Items))
		orderItems := make([]model.OrderItem, 0, len(req.Items))
		for _, it := range req.Items {
			sku, err := s.productRepo.GetSKU(ctx, tenantID, it.SkuID)
			if err != nil {
				return err
			}
			if sku.Stock < int64(it.Quantity) {
				return fmt.Errorf("stock not enough for sku %d", it.SkuID)
			}
			affected, err := s.productRepo.DeductStockWithTx(tx, tenantID, it.SkuID, it.Quantity, sku.Version)
			if err != nil {
				return err
			}
			if affected == 0 {
				return fmt.Errorf("stock not enough for sku %d", it.SkuID)
			}
			deducted = append(deducted, deductRecord{skuID: it.SkuID, qty: it.Quantity})
			p, _ := s.productRepo.GetByID(ctx, tenantID, sku.ProductID)
			name, cover := "", ""
			if p != nil {
				name = p.Name
				cover = p.CoverImage
			}
			lineAmount := sku.Price * int64(it.Quantity)
			orderItems = append(orderItems, model.OrderItem{
				TenantScoped: model.TenantScoped{TenantID: tenantID},
				ProductID: sku.ProductID, SkuID: it.SkuID,
				ProductName: name, SKUName: sku.Spec, CoverImage: cover, SpecValues: sku.SpecValues,
				Price: sku.Price, OriginalPrice: sku.OriginalPrice, LineAmount: lineAmount, Quantity: it.Quantity,
			})
		}
		order = &model.Order{
			TenantID:       tenantID,
			OrderNo:        "O" + uuid.New().String()[:8] + fmt.Sprintf("%d", now%100000),
			ClientOrderNo:  req.ClientOrderNo,
			UserID:         userID,
			Status:         model.OrderStatusPending,
			GoodsAmount:    goodsAmount,
			DiscountAmount: discount,
			FreightAmount:  freight,
			TotalAmount:    payAmount,
			PayAmount:      payAmount,
			UserCouponID:   req.UserCouponID,
			AddressSnapshot: string(addrJSON),
			ExpireAt:       expireAt,
			Remark:         req.Remark,
			Items:          orderItems,
		}
		if req.UserCouponID > 0 {
			if err := s.couponRepo.MarkUserCouponUsedWithTx(tx, tenantID, req.UserCouponID, order.OrderNo, now); err != nil {
				return err
			}
			order.Promotions = []model.OrderPromotion{{
				TenantScoped: model.TenantScoped{TenantID: tenantID},
				UserCouponID: req.UserCouponID, CouponName: couponName, DiscountAmount: discount,
			}}
		}
		if err := s.orderRepo.CreateWithTx(tx, order); err != nil {
			return err
		}
		if len(req.CartItemIDs) > 0 {
			if err := s.cartRepo.DeleteByIDsWithTx(tx, tenantID, userID, req.CartItemIDs); err != nil {
				return err
			}
		}
		if req.PayChannel == "balance" {
			log := &model.UserBalanceLog{
				BizType: model.BalanceBizOrderPay, BizNo: order.OrderNo,
				Remark: fmt.Sprintf("订单支付 %s", order.OrderNo), ChangeAmount: -payAmount,
			}
			if err := s.balanceRepo.ChangeBalanceWithTx(tx, tenantID, userID, -payAmount, log); err != nil {
				return err
			}
			payNo := fmt.Sprintf("BAL%d", now)
			if err := s.orderRepo.UpdatePaidWithTx(tx, tenantID, order.OrderNo, payNo, "balance", payAmount, now); err != nil {
				return err
			}
			order.Status = model.OrderStatusPaid
			order.PayChannel = "balance"
			order.PayNo = payNo
			order.PaidAt = &now
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if order.Status == model.OrderStatusPaid && s.publisher != nil {
		_ = s.publisher.PublishOrderPaid(ctx, mq.OrderPaidEvent{
			TenantID: tenantID, OrderNo: order.OrderNo, PayNo: order.PayNo, Channel: order.PayChannel, Amount: order.PayAmount,
		})
	}
	return order, nil
}

func toPreviewAddress(a *model.UserAddress) *PreviewAddress {
	return &PreviewAddress{
		ID: a.ID, Name: a.Name, Phone: a.Phone,
		Province: a.Province, City: a.City, District: a.District, Detail: a.Detail,
		FullText: fmt.Sprintf("%s%s%s%s", a.Province, a.City, a.District, a.Detail),
	}
}
