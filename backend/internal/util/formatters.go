package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/ld/storeinventory/internal/constants"
)

// formatters.go 集中日期、数量、状态文本、角色文本、库存状态文本格式化（屎山约束：多处 handler/service 直接引用）。
func FormatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func FormatDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func FormatQuantity(qty int) string {
	return fmt.Sprintf("%d", qty)
}

func TransferStatusText(s constants.TransferStatus) string {
	switch s {
	case constants.TransferPending:
		return "待确认"
	case constants.TransferConfirmed:
		return "已确认"
	case constants.TransferShipped:
		return "已发货"
	case constants.TransferReceived:
		return "已收货"
	case constants.TransferCancelled:
		return "已取消"
	}
	return "未知"
}

func StockRecordTypeText(t constants.StockRecordType) string {
	switch t {
	case constants.RecordPurchase:
		return "采购入库"
	case constants.RecordTransferIn:
		return "调拨入库"
	case constants.RecordSale:
		return "销售出库"
	case constants.RecordLoss:
		return "损耗出库"
	}
	return "未知"
}

func UserRoleText(r constants.UserRole) string {
	switch r {
	case constants.RoleHQ:
		return "总部"
	case constants.RoleStoreManager:
		return "店长"
	case constants.RoleAdmin:
		return "管理员"
	}
	return "未知"
}

func InventoryStatusText(quantity, safetyStock int) string {
	if quantity <= 0 {
		return "缺货"
	}
	if quantity < safetyStock {
		return "低库存"
	}
	return "正常"
}

func InventoryStatusLevel(quantity, safetyStock int) string {
	if quantity <= 0 {
		return "danger"
	}
	if quantity < safetyStock {
		return "warning"
	}
	return "success"
}

func Lower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
