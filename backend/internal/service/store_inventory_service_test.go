package service

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/util"
)

type mockInventoryRepo struct {
	items map[string]*model.StoreInventory
	seq   uint
}

func newMockInventoryRepo() *mockInventoryRepo {
	return &mockInventoryRepo{items: make(map[string]*model.StoreInventory)}
}

func keyOf(storeID, skuID uint) string {
	return string(rune(storeID)) + ":" + string(rune(skuID))
}

func (m *mockInventoryRepo) Create(inv *model.StoreInventory) error { return m.CreateTx(nil, inv) }
func (m *mockInventoryRepo) CreateTx(tx *gorm.DB, inv *model.StoreInventory) error {
	k := keyOf(inv.StoreID, inv.SKUID)
	if _, ok := m.items[k]; ok {
		return util.ErrConflict
	}
	m.seq++
	inv.ID = m.seq
	m.items[k] = inv
	return nil
}
func (m *mockInventoryRepo) FindByStoreAndSKU(storeID, skuID uint) (*model.StoreInventory, error) {
	return m.FindByStoreAndSKUTx(nil, storeID, skuID)
}
func (m *mockInventoryRepo) FindByStoreAndSKUTx(tx *gorm.DB, storeID, skuID uint) (*model.StoreInventory, error) {
	if inv, ok := m.items[keyOf(storeID, skuID)]; ok {
		cp := *inv
		return &cp, nil
	}
	return nil, util.ErrNotFound
}
func (m *mockInventoryRepo) FindByID(id uint) (*model.StoreInventory, error) {
	return nil, util.ErrNotFound
}
func (m *mockInventoryRepo) List(page, pageSize int, storeID, skuID uint) ([]model.StoreInventory, int64, error) {
	return nil, 0, nil
}
func (m *mockInventoryRepo) ListAlerts() ([]model.StoreInventory, error) { return nil, nil }
func (m *mockInventoryRepo) Update(inv *model.StoreInventory) error      { return nil }
func (m *mockInventoryRepo) AdjustQuantity(storeID, skuID uint, delta int) error {
	return m.AdjustQuantityTx(nil, storeID, skuID, delta)
}
func (m *mockInventoryRepo) AdjustQuantityTx(tx *gorm.DB, storeID, skuID uint, delta int) error {
	inv, err := m.FindByStoreAndSKUTx(tx, storeID, skuID)
	if err != nil {
		return err
	}
	inv.Quantity += delta
	m.items[keyOf(storeID, skuID)] = inv
	return nil
}

type mockSKURepo struct{}

func (m *mockSKURepo) Create(sku *model.SKU) error { return nil }
func (m *mockSKURepo) FindByID(id uint) (*model.SKU, error) {
	return &model.SKU{ID: id}, nil
}
func (m *mockSKURepo) FindByCode(code string) (*model.SKU, error) { return nil, util.ErrNotFound }
func (m *mockSKURepo) List(page, pageSize int, category, keyword string) ([]model.SKU, int64, error) {
	return nil, 0, nil
}
func (m *mockSKURepo) BatchCreate(skus []*model.SKU) (int, error) { return len(skus), nil }
func (m *mockSKURepo) Update(sku *model.SKU) error                { return nil }
func (m *mockSKURepo) Delete(id uint) error                       { return nil }

func newTestInventoryService() (StoreInventoryService, *mockInventoryRepo) {
	repo := newMockInventoryRepo()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewStoreInventoryService(repo, &mockSKURepo{}, nil, logger), repo
}

func TestStoreInventoryCheckSufficient(t *testing.T) {
	svc, repo := newTestInventoryService()
	if _, err := svc.Ensure(1, 10, 100); err != nil {
		t.Fatalf("ensure failed: %v", err)
	}
	tests := []struct {
		name    string
		qty     int
		wantErr bool
	}{
		{"sufficient", 100, false},
		{"more than stock", 101, true},
		{"zero", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.CheckSufficient(1, 10, tt.qty)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckSufficient(%d) err=%v, wantErr=%v", tt.qty, err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, util.ErrStockNotEnough) {
				t.Fatalf("expected ErrStockNotEnough, got %v", err)
			}
		})
	}
	_ = repo
}

func TestStoreInventoryAdjustQuantity(t *testing.T) {
	svc, _ := newTestInventoryService()
	if _, err := svc.Ensure(1, 20, 10); err != nil {
		t.Fatalf("ensure failed: %v", err)
	}
	tests := []struct {
		name    string
		delta   int
		wantQty int
		wantErr bool
	}{
		{"add stock", 5, 15, false},
		{"deduct stock", -3, 12, false},
		{"deduct too much", -100, 12, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.AdjustQuantity(1, 20, tt.delta)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AdjustQuantity(%d) err=%v, wantErr=%v", tt.delta, err, tt.wantErr)
			}
			inv, getErr := svc.GetByStoreAndSKU(1, 20)
			if getErr != nil {
				t.Fatalf("get inventory failed: %v", getErr)
			}
			if inv.Quantity != tt.wantQty {
				t.Fatalf("quantity=%d, want %d", inv.Quantity, tt.wantQty)
			}
		})
	}
}

func TestStockDirection(t *testing.T) {
	tests := []struct {
		typ  constants.StockRecordType
		want int
	}{
		{constants.RecordPurchase, 1},
		{constants.RecordTransferIn, 1},
		{constants.RecordSale, -1},
		{constants.RecordLoss, -1},
	}
	for _, tt := range tests {
		if got := constants.StockDirection(tt.typ); got != tt.want {
			t.Errorf("StockDirection(%s)=%d, want %d", tt.typ, got, tt.want)
		}
	}
}
