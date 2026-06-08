package service

import (
	"context"
	"fmt"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
)

type BalanceLogVO struct {
	ID           uint64 `json:"id"`
	UserID       uint64 `json:"user_id"`
	ChangeAmount int64  `json:"change_amount"`
	BalanceAfter int64  `json:"balance_after"`
	BizType      string `json:"biz_type"`
	BizNo        string `json:"biz_no"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"created_at"`
	OrderNo      string `json:"order_no"`
	GoodsSummary string `json:"goods_summary"`
}

func enrichBalanceLogs(ctx context.Context, orderRepo *repository.OrderRepo, tenantID uint64, logs []model.UserBalanceLog) ([]BalanceLogVO, error) {
	orderNos := make([]string, 0)
	seen := map[string]struct{}{}
	for _, log := range logs {
		if !isOrderBalanceBiz(log.BizType) || log.BizNo == "" {
			continue
		}
		if _, ok := seen[log.BizNo]; ok {
			continue
		}
		seen[log.BizNo] = struct{}{}
		orderNos = append(orderNos, log.BizNo)
	}
	summaryMap, err := buildOrderGoodsSummaryMap(ctx, orderRepo, tenantID, orderNos)
	if err != nil {
		return nil, err
	}
	vos := make([]BalanceLogVO, 0, len(logs))
	for _, log := range logs {
		vo := BalanceLogVO{
			ID:           log.ID,
			UserID:       log.UserID,
			ChangeAmount: log.ChangeAmount,
			BalanceAfter: log.BalanceAfter,
			BizType:      log.BizType,
			BizNo:        log.BizNo,
			Remark:       log.Remark,
			CreatedAt:    log.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if isOrderBalanceBiz(log.BizType) && log.BizNo != "" {
			vo.OrderNo = log.BizNo
			vo.GoodsSummary = summaryMap[log.BizNo]
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

func isOrderBalanceBiz(bizType string) bool {
	return bizType == model.BalanceBizOrderPay || bizType == model.BalanceBizOrderRefund
}

func formatGoodsSummary(items []model.OrderItem) string {
	if len(items) == 0 {
		return ""
	}
	name := items[0].ProductName
	if name == "" {
		name = "商品"
	}
	totalQty := 0
	for _, it := range items {
		totalQty += it.Quantity
	}
	if len(items) == 1 {
		if items[0].Quantity > 1 {
			return fmt.Sprintf("%s x%d", name, items[0].Quantity)
		}
		return name
	}
	others := len(items) - 1
	if others > 0 {
		return fmt.Sprintf("%s 等%d件", name, totalQty)
	}
	return name
}

func buildOrderGoodsSummaryMap(ctx context.Context, orderRepo *repository.OrderRepo, tenantID uint64, orderNos []string) (map[string]string, error) {
	result := make(map[string]string, len(orderNos))
	orders, err := orderRepo.ListByOrderNos(ctx, tenantID, orderNos)
	if err != nil {
		return nil, err
	}
	for _, o := range orders {
		result[o.OrderNo] = formatGoodsSummary(o.Items)
	}
	return result, nil
}
