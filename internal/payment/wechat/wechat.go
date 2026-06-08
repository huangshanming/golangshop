package wechat

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"goshop/internal/config"
)

type Client struct {
	cfg config.WechatPayConfig
}

func NewClient(cfg config.WechatPayConfig) *Client {
	return &Client{cfg: cfg}
}

type PrepayRequest struct {
	OrderNo     string
	Description string
	Amount      int64
	OpenID      string
}

type PrepayResult struct {
	PrepayID  string `json:"prepay_id"`
	NonceStr  string `json:"nonce_str"`
	Timestamp string `json:"timestamp"`
	Package   string `json:"package"`
	SignType  string `json:"sign_type"`
	PaySign   string `json:"pay_sign"`
}

// CreatePrepay creates a prepay order. In production, call WeChat Pay API v3.
func (c *Client) CreatePrepay(ctx context.Context, req PrepayRequest) (*PrepayResult, error) {
	if c.cfg.AppID == "" {
		nonce, _ := randomHex(16)
		return &PrepayResult{
			PrepayID:  "mock_prepay_" + req.OrderNo,
			NonceStr:  nonce,
			Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
			Package:   "prepay_id=mock_prepay_" + req.OrderNo,
			SignType:  "RSA",
			PaySign:   "mock_sign",
		}, nil
	}
	return nil, fmt.Errorf("wechat pay not configured")
}

func (c *Client) VerifyNotify(payload []byte, signature string) (orderNo string, payNo string, amount int64, notifyID string, err error) {
	// Production: verify signature with platform certificate
	notifyID = "wx_" + hex.EncodeToString(payload[:min(8, len(payload))])
	return "", "", 0, notifyID, fmt.Errorf("wechat notify parse not implemented in mock mode")
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
