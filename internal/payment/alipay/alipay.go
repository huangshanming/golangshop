package alipay

import (
	"context"
	"fmt"
	"time"

	"goshop/internal/config"
)

type Client struct {
	cfg config.AlipayConfig
}

func NewClient(cfg config.AlipayConfig) *Client {
	return &Client{cfg: cfg}
}

type PrepayRequest struct {
	OrderNo     string
	Subject     string
	Amount      int64
	ReturnURL   string
}

type PrepayResult struct {
	OrderString string `json:"order_string"`
}

func (c *Client) CreatePrepay(ctx context.Context, req PrepayRequest) (*PrepayResult, error) {
	if c.cfg.AppID == "" {
		return &PrepayResult{
			OrderString: fmt.Sprintf("mock_alipay_order_%s_%d", req.OrderNo, time.Now().Unix()),
		}, nil
	}
	return nil, fmt.Errorf("alipay not configured")
}

func (c *Client) VerifyNotify(values map[string]string) (orderNo, payNo string, amount int64, notifyID string, err error) {
	notifyID = values["notify_id"]
	if notifyID == "" {
		notifyID = "alipay_" + values["out_trade_no"]
	}
	orderNo = values["out_trade_no"]
	payNo = values["trade_no"]
	return orderNo, payNo, 0, notifyID, nil
}
