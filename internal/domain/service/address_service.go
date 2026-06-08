package service

import (
	"context"
	"fmt"
	"time"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
)

type AddressService struct {
	repo *repository.AddressRepo
}

func NewAddressService(repo *repository.AddressRepo) *AddressService {
	return &AddressService{repo: repo}
}

type AddressReq struct {
	Name      string `json:"name" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Detail    string `json:"detail" binding:"required"`
	IsDefault bool   `json:"is_default"`
}

func (s *AddressService) List(ctx context.Context, tenantID, userID uint64) ([]model.UserAddress, error) {
	return s.repo.List(ctx, tenantID, userID)
}

func (s *AddressService) Create(ctx context.Context, tenantID, userID uint64, req AddressReq) (*model.UserAddress, error) {
	a := &model.UserAddress{
		TenantScoped: model.TenantScoped{TenantID: tenantID},
		UserID: userID, Name: req.Name, Phone: req.Phone,
		Province: req.Province, City: req.City, District: req.District, Detail: req.Detail,
		IsDefault: req.IsDefault,
	}
	return a, s.repo.Create(ctx, a)
}

func (s *AddressService) Update(ctx context.Context, tenantID, userID, id uint64, req AddressReq) error {
	a, err := s.repo.GetByID(ctx, tenantID, userID, id)
	if err != nil {
		return err
	}
	a.Name = req.Name
	a.Phone = req.Phone
	a.Province = req.Province
	a.City = req.City
	a.District = req.District
	a.Detail = req.Detail
	a.IsDefault = req.IsDefault
	return s.repo.Update(ctx, a)
}

func (s *AddressService) Delete(ctx context.Context, tenantID, userID, id uint64) error {
	return s.repo.Delete(ctx, tenantID, userID, id)
}

func (s *AddressService) GetDefault(ctx context.Context, tenantID, userID uint64) (*model.UserAddress, error) {
	return s.repo.GetDefault(ctx, tenantID, userID)
}

type CouponService struct {
	repo *repository.CouponRepo
}

func NewCouponService(repo *repository.CouponRepo) *CouponService {
	return &CouponService{repo: repo}
}

type CouponVO struct {
	UserCouponID   uint64 `json:"user_coupon_id"`
	CouponID       uint64 `json:"coupon_id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	ThresholdAmount int64 `json:"threshold_amount"`
	DiscountValue  int64  `json:"discount_value"`
	MaxDiscount    int64  `json:"max_discount"`
	ValidTo        int64  `json:"valid_to"`
	CanUse         bool   `json:"can_use"`
	Reason         string `json:"reason"`
	DiscountAmount int64  `json:"discount_amount"`
}

func (s *CouponService) ListAvailable(ctx context.Context, tenantID, userID uint64, goodsAmount int64) ([]CouponVO, error) {
	list, err := s.repo.ListUserCoupons(ctx, tenantID, userID, model.UserCouponUnused)
	if err != nil {
		return nil, err
	}
	vos := make([]CouponVO, 0, len(list))
	for _, uc := range list {
		coupon, err := s.repo.GetByID(ctx, tenantID, uc.CouponID)
		if err != nil {
			continue
		}
		vo := CouponVO{
			UserCouponID: uc.ID, CouponID: coupon.ID, Name: coupon.Name, Type: coupon.Type,
			ThresholdAmount: coupon.ThresholdAmount, DiscountValue: coupon.DiscountValue,
			MaxDiscount: coupon.MaxDiscount, ValidTo: coupon.ValidTo,
		}
		now := time.Now().Unix()
		if coupon.Status != 1 || now < coupon.ValidFrom || now > coupon.ValidTo {
			vo.CanUse = false
			vo.Reason = "已过期"
		} else if goodsAmount < coupon.ThresholdAmount {
			vo.CanUse = false
			vo.Reason = "未满足门槛"
		} else {
			vo.CanUse = true
			vo.DiscountAmount = calcCouponDiscount(coupon, goodsAmount)
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

func (s *CouponService) Receive(ctx context.Context, tenantID, userID, couponID uint64) error {
	coupon, err := s.repo.GetByID(ctx, tenantID, couponID)
	if err != nil {
		return err
	}
	count, _ := s.repo.CountUserCoupon(ctx, tenantID, userID, couponID)
	if int(count) >= coupon.PerUserLimit {
		return fmt.Errorf("coupon receive limit reached")
	}
	now := time.Now().Unix()
	uc := &model.UserCoupon{
		TenantScoped: model.TenantScoped{TenantID: tenantID},
		UserID: userID, CouponID: couponID, Status: model.UserCouponUnused, ReceivedAt: now,
	}
	return s.repo.CreateUserCoupon(ctx, uc)
}

func (s *CouponService) ListMine(ctx context.Context, tenantID, userID uint64) ([]CouponVO, error) {
	return s.ListAvailable(ctx, tenantID, userID, 0)
}

type AdminCouponReq struct {
	Name            string `json:"name" binding:"required"`
	Type            string `json:"type" binding:"required"`
	ThresholdAmount int64  `json:"threshold_amount"`
	DiscountValue   int64  `json:"discount_value" binding:"required"`
	MaxDiscount     int64  `json:"max_discount"`
	ValidFrom       int64  `json:"valid_from" binding:"required"`
	ValidTo         int64  `json:"valid_to" binding:"required"`
	TotalCount      int64  `json:"total_count"`
	PerUserLimit    int    `json:"per_user_limit"`
	Status          int8   `json:"status"`
}

func (s *CouponService) ListAdmin(ctx context.Context, tenantID uint64, page, pageSize int, keyword string) ([]model.Coupon, int64, error) {
	return s.repo.ListPage(ctx, tenantID, page, pageSize, keyword)
}

func (s *CouponService) GetAdmin(ctx context.Context, tenantID, id uint64) (*model.Coupon, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *CouponService) CreateAdmin(ctx context.Context, tenantID uint64, req AdminCouponReq) (*model.Coupon, error) {
	if req.Type != model.CouponTypeFixed && req.Type != model.CouponTypePercent {
		return nil, fmt.Errorf("invalid coupon type")
	}
	if req.PerUserLimit < 1 {
		req.PerUserLimit = 1
	}
	if req.Status == 0 {
		req.Status = 1
	}
	c := &model.Coupon{
		TenantScoped:    model.TenantScoped{TenantID: tenantID},
		Name:            req.Name,
		Type:            req.Type,
		ThresholdAmount: req.ThresholdAmount,
		DiscountValue:   req.DiscountValue,
		MaxDiscount:     req.MaxDiscount,
		ValidFrom:       req.ValidFrom,
		ValidTo:         req.ValidTo,
		TotalCount:      req.TotalCount,
		PerUserLimit:    req.PerUserLimit,
		Status:          req.Status,
	}
	return c, s.repo.Create(ctx, c)
}

func (s *CouponService) UpdateAdmin(ctx context.Context, tenantID, id uint64, req AdminCouponReq) (*model.Coupon, error) {
	c, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if req.Type != model.CouponTypeFixed && req.Type != model.CouponTypePercent {
		return nil, fmt.Errorf("invalid coupon type")
	}
	c.Name = req.Name
	c.Type = req.Type
	c.ThresholdAmount = req.ThresholdAmount
	c.DiscountValue = req.DiscountValue
	c.MaxDiscount = req.MaxDiscount
	c.ValidFrom = req.ValidFrom
	c.ValidTo = req.ValidTo
	c.TotalCount = req.TotalCount
	if req.PerUserLimit >= 1 {
		c.PerUserLimit = req.PerUserLimit
	}
	c.Status = req.Status
	return c, s.repo.Update(ctx, c)
}

func (s *CouponService) UpdateStatusAdmin(ctx context.Context, tenantID, id uint64, status int8) error {
	return s.repo.UpdateStatus(ctx, tenantID, id, status)
}

type BalanceService struct {
	repo      *repository.BalanceRepo
	userRepo  *repository.UserRepo
	orderRepo *repository.OrderRepo
}

func NewBalanceService(repo *repository.BalanceRepo, userRepo *repository.UserRepo, orderRepo *repository.OrderRepo) *BalanceService {
	return &BalanceService{repo: repo, userRepo: userRepo, orderRepo: orderRepo}
}

func (s *BalanceService) Adjust(ctx context.Context, tenantID, userID uint64, change int64, remark string, operatorID uint64) error {
	log := &model.UserBalanceLog{
		BizType: model.BalanceBizAdminAdjust, BizNo: fmt.Sprintf("ADJ%d", time.Now().UnixNano()),
		Remark: remark, ChangeAmount: change, OperatorID: operatorID,
	}
	return s.repo.ChangeBalance(ctx, tenantID, userID, change, log)
}

func (s *BalanceService) ListLogs(ctx context.Context, tenantID, userID uint64, page, pageSize int) ([]BalanceLogVO, int64, error) {
	logs, total, err := s.repo.ListLogs(ctx, tenantID, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	vos, err := enrichBalanceLogs(ctx, s.orderRepo, tenantID, logs)
	if err != nil {
		return nil, 0, err
	}
	return vos, total, nil
}

func (s *BalanceService) ListUsers(ctx context.Context, tenantID uint64, page, pageSize int, keyword string) ([]model.User, int64, error) {
	return s.userRepo.List(ctx, tenantID, page, pageSize, keyword)
}
