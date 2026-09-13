package router

import (
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
)

// seedData 初始化种子数据：门店、用户、SKU、门店库存。
func seedData(db *gorm.DB, logger *slog.Logger) error {
	hash := func(pwd string) string {
		h, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("seed hash failed", "error", err)
			return ""
		}
		return string(h)
	}

	stores := []model.Store{
		{Code: "ST001", Name: "北京朝阳门店", Address: "北京市朝阳区建国路 88 号"},
		{Code: "ST002", Name: "上海静安门店", Address: "上海市静安区南京西路 1266 号"},
		{Code: "ST003", Name: "广州天河门店", Address: "广州市天河区天河路 208 号"},
	}
	if err := db.Create(&stores).Error; err != nil {
		return fmt.Errorf("seed stores: %w", err)
	}

	users := []model.User{
		{Username: "admin", PasswordHash: hash("admin123"), Name: "系统管理员", Role: constants.RoleAdmin},
		{Username: "hquser", PasswordHash: hash("hq123456"), Name: "总部采购", Role: constants.RoleHQ},
		{Username: "manager1", PasswordHash: hash("sm123456"), Name: "朝阳店长", Role: constants.RoleStoreManager, StoreID: &stores[0].ID},
		{Username: "manager2", PasswordHash: hash("sm123456"), Name: "静安店长", Role: constants.RoleStoreManager, StoreID: &stores[1].ID},
	}
	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("seed users: %w", err)
	}

	skus := []model.SKU{
		{Code: "SKU1001", Name: "可口可乐 330ml", Spec: "330ml/罐", Barcode: "6902538001000", Category: "饮料", Unit: "罐"},
		{Code: "SKU1002", Name: "农夫山泉 550ml", Spec: "550ml/瓶", Barcode: "6902538002000", Category: "饮料", Unit: "瓶"},
		{Code: "SKU2001", Name: "康师傅红烧牛肉面", Spec: "袋装 105g", Barcode: "6902538003000", Category: "方便食品", Unit: "袋"},
		{Code: "SKU2002", Name: "乐事薯片原味", Spec: "70g/袋", Barcode: "6902538004000", Category: "零食", Unit: "袋"},
		{Code: "SKU3001", Name: "维达抽纸 3 层", Spec: "120 抽/包", Barcode: "6902538005000", Category: "日用品", Unit: "包"},
	}
	if err := db.Create(&skus).Error; err != nil {
		return fmt.Errorf("seed skus: %w", err)
	}

	inventories := []model.StoreInventory{
		{StoreID: stores[0].ID, SKUID: skus[0].ID, Quantity: 120, SafetyStock: 50},
		{StoreID: stores[0].ID, SKUID: skus[1].ID, Quantity: 20, SafetyStock: 80},
		{StoreID: stores[0].ID, SKUID: skus[2].ID, Quantity: 200, SafetyStock: 60},
		{StoreID: stores[1].ID, SKUID: skus[0].ID, Quantity: 30, SafetyStock: 40},
		{StoreID: stores[1].ID, SKUID: skus[3].ID, Quantity: 150, SafetyStock: 50},
		{StoreID: stores[2].ID, SKUID: skus[4].ID, Quantity: 8, SafetyStock: 30},
	}
	if err := db.Create(&inventories).Error; err != nil {
		return fmt.Errorf("seed inventories: %w", err)
	}

	records := []model.StockRecord{
		{StoreID: stores[0].ID, SKUID: skus[0].ID, RecordType: constants.RecordPurchase, Quantity: 200},
		{StoreID: stores[1].ID, SKUID: skus[0].ID, RecordType: constants.RecordPurchase, Quantity: 150},
		{StoreID: stores[0].ID, SKUID: skus[1].ID, RecordType: constants.RecordSale, Quantity: 30},
	}
	if err := db.Create(&records).Error; err != nil {
		return fmt.Errorf("seed stock records: %w", err)
	}
	logger.Info("seed data created", "stores", len(stores), "users", len(users), "skus", len(skus))
	return nil
}
