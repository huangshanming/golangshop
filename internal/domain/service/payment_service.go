package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"goshop/internal/config"
	"goshop/internal/domain/model"
	"goshop/internal/payment/alipay"
	"goshop/internal/payment/wechat"
	"goshop/internal/repository"
)

type PaymentService struct {
	orderSvc    *OrderService
	balanceRepo *repository.BalanceRepo
	notifyRepo  *repository.PayNotifyRepo
	wechat      *wechat.Client
	alipay      *alipay.Client
	wechatCfg   config.WechatPayConfig
	alipayCfg   config.AlipayConfig
}

func NewPaymentService(
	orderSvc *OrderService,
	balanceRepo *repository.BalanceRepo,
	notifyRepo *repository.PayNotifyRepo,
	wechatClient *wechat.Client,
	alipayClient *alipay.Client,
	paymentCfg config.PaymentConfig,
) *PaymentService {
	return &PaymentService{
		orderSvc: orderSvc, balanceRepo: balanceRepo, notifyRepo: notifyRepo,
		wechat: wechatClient, alipay: alipayClient,
		wechatCfg: paymentCfg.Wechat, alipayCfg: paymentCfg.Alipay,
	}
}

type PayReq struct {
	OrderNo string `json:"order_no" binding:"required"`
	Channel string `json:"channel" binding:"required,oneof=balance wechat alipay"`
	OpenID  string `json:"openid"`
}

func (s *PaymentService) CreatePay(ctx context.Context, tenantID, userID uint64, req PayReq) (interface{}, error) {
	order, err := s.orderSvc.GetInternal(ctx, tenantID, req.OrderNo)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != model.OrderStatusPending {
		return nil, fmt.Errorf("order status invalid")
	}
	amount := order.TotalAmount
	switch req.Channel {
	case "balance":
		log := &model.UserBalanceLog{
			BizType: model.BalanceBizOrderPay, BizNo: order.OrderNo,
			Remark: "订单支付", ChangeAmount: -amount,
		}
		if err := s.balanceRepo.ChangeBalance(ctx, tenantID, userID, -amount, log); err != nil {
			return nil, err
		}
		payNo := fmt.Sprintf("BAL%d", time.Now().UnixNano())
		if err := s.orderSvc.MarkPaid(ctx, tenantID, order.OrderNo, payNo, "balance", amount); err != nil {
			return nil, err
		}
		return map[string]interface{}{"paid": true, "channel": "balance"}, nil
	case "wechat":
		return s.wechat.CreatePrepay(ctx, wechat.PrepayRequest{
			OrderNo: order.OrderNo, Description: "GoShop Order", Amount: amount, OpenID: req.OpenID,
		})
	case "alipay":
		return s.alipay.CreatePrepay(ctx, alipay.PrepayRequest{
			OrderNo: order.OrderNo, Subject: "GoShop Order", Amount: amount,
		})
	default:
		return nil, fmt.Errorf("unsupported channel")
	}
}

func (s *PaymentService) HandleWechatNotify(ctx context.Context, tenantID uint64, body []byte, signature string) error {
	orderNo, payNo, amount, notifyID, err := s.wechat.VerifyNotify(body, signature)
	if err != nil {
		// mock mode: parse JSON body
		var mock struct {
			OrderNo string `json:"order_no"`
			PayNo   string `json:"pay_no"`
			Amount  int64  `json:"amount"`
		}
		if json.Unmarshal(body, &mock) == nil && mock.OrderNo != "" {
			orderNo, payNo, amount = mock.OrderNo, mock.PayNo, mock.Amount
			notifyID = "wx_mock_" + mock.PayNo
		} else {
			return err
		}
	}
	return s.processNotify(ctx, tenantID, "wechat", notifyID, orderNo, payNo, amount, string(body))
}

func (s *PaymentService) HandleAlipayNotify(ctx context.Context, tenantID uint64, values map[string]string) error {
	orderNo, payNo, amount, notifyID, err := s.alipay.VerifyNotify(values)
	if err != nil {
		return err
	}
	if amount == 0 {
		order, _ := s.orderSvc.GetInternal(ctx, tenantID, orderNo)
		if order != nil {
			amount = order.TotalAmount
		}
	}
	payload, _ := json.Marshal(values)
	return s.processNotify(ctx, tenantID, "alipay", notifyID, orderNo, payNo, amount, string(payload))
}

func (s *PaymentService) processNotify(ctx context.Context, tenantID uint64, channel, notifyID, orderNo, payNo string, amount int64, payload string) error {
	exists, err := s.notifyRepo.Exists(ctx, notifyID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	log := &model.PayNotifyLog{
		TenantID: tenantID, Channel: channel, NotifyID: notifyID,
		OrderNo: orderNo, Payload: payload,
	}
	if err := s.notifyRepo.Create(ctx, log); err != nil {
		return err
	}
	if err := s.orderSvc.MarkPaid(ctx, tenantID, orderNo, payNo, channel, amount); err != nil {
		return err
	}
	return s.notifyRepo.MarkProcessed(ctx, notifyID)
}

// MockPay simulates successful payment for development.
func (s *PaymentService) MockPay(ctx context.Context, tenantID uint64, orderNo string) error {
	payNo := fmt.Sprintf("MOCK%d", time.Now().UnixNano())
	order, err := s.orderSvc.GetInternal(ctx, tenantID, orderNo)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]interface{}{"order_no": orderNo, "pay_no": payNo, "amount": order.TotalAmount})
	return s.HandleWechatNotify(ctx, tenantID, body, "")
}
