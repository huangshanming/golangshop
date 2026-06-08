package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const previewTokenTTL = 15 * time.Minute

type previewPayload struct {
	TenantID      uint64            `json:"tid"`
	UserID        uint64            `json:"uid"`
	AddressID     uint64            `json:"aid"`
	UserCouponID  uint64            `json:"ucid"`
	Items         []previewItemSign `json:"items"`
	GoodsAmount   int64             `json:"ga"`
	DiscountAmt   int64             `json:"da"`
	FreightAmt    int64             `json:"fa"`
	PayAmount     int64             `json:"pa"`
	ExpiresAt     int64             `json:"exp"`
}

type previewItemSign struct {
	SkuID    uint64 `json:"sku_id"`
	Quantity int    `json:"qty"`
	Price    int64  `json:"price"`
}

func signPreview(secret string, p previewPayload) (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(b)
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	payload := base64.RawURLEncoding.EncodeToString(b)
	return payload + "." + sig, nil
}

func verifyPreview(secret, token string, tenantID, userID uint64) (*previewPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid preview token")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return nil, errors.New("preview token signature mismatch")
	}
	var p previewPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	if p.TenantID != tenantID || p.UserID != userID {
		return nil, errors.New("preview token user mismatch")
	}
	if time.Now().Unix() > p.ExpiresAt {
		return nil, errors.New("preview token expired")
	}
	return &p, nil
}

func itemsSignKey(items []previewItemSign) string {
	cp := append([]previewItemSign(nil), items...)
	sort.Slice(cp, func(i, j int) bool {
		if cp[i].SkuID == cp[j].SkuID {
			return cp[i].Quantity < cp[j].Quantity
		}
		return cp[i].SkuID < cp[j].SkuID
	})
	parts := make([]string, 0, len(cp))
	for _, it := range cp {
		parts = append(parts, fmt.Sprintf("%d:%d:%d", it.SkuID, it.Quantity, it.Price))
	}
	return strings.Join(parts, "|")
}

func parseOrderExpireMinutes() int64 { return 30 }

func formatDisplayStatus(status int8, reviewStatus int8, hasAfterSale bool, afterSaleType string) string {
	if hasAfterSale {
		switch afterSaleType {
		case "exchange":
			return "售后中-换货"
		case "return_refund":
			return "售后中-退货退款"
		case "refund_only":
			return "售后中-仅退款"
		default:
			return "售后中"
		}
	}
	switch status {
	case 0:
		return "待付款"
	case 1:
		return "待发货"
	case 2:
		return "待收货"
	case 3:
		if reviewStatus == 1 {
			return "待评价"
		}
		if reviewStatus == 2 {
			return "已评价"
		}
		return "已完成"
	case 4:
		return "已取消"
	case 5:
		return "已关闭"
	default:
		return "未知"
	}
}

func yuanHint(cent int64) string {
	return strconv.FormatFloat(float64(cent)/100, 'f', 2, 64)
}
