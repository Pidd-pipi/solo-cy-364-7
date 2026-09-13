package service

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/repository"
	"github.com/ld/storeinventory/internal/util"
)

// StockRecordService 出入库记录业务逻辑。
type StockRecordService interface {
	Create(storeID, skuID uint, recordType constants.StockRecordType, quantity int, relatedOrderID *uint) (*model.StockRecord, error)
	CreateTx(tx *gorm.DB, storeID, skuID uint, recordType constants.StockRecordType, quantity int, relatedOrderID *uint) (*model.StockRecord, error)
	List(page, pageSize int, storeID, skuID uint, recordType constants.StockRecordType) ([]model.StockRecord, int64, error)
	Export(storeID uint) ([]model.StockRecord, error)
	CreateStocktake(storeID, skuID uint, stocktakeDate string, actualQty int, remark string) (*model.Stocktake, error)
	ListStocktakes(page, pageSize int, storeID uint) ([]model.Stocktake, int64, error)
	ReplenishSuggestions() ([]util.ReplenishSuggestion, error)
}

type stockRecordService struct {
	recordRepo repository.StockRecordRepository
	invRepo    repository.StoreInventoryRepository
	storeRepo  repository.StoreRepository
	skuRepo    repository.SKURepository
	invSvc     StoreInventoryService
	db         *gorm.DB
	logger     *slog.Logger
}

// NewStockRecordService 构造出入库记录服务。
func NewStockRecordService(recordRepo repository.StockRecordRepository, invRepo repository.StoreInventoryRepository, storeRepo repository.StoreRepository, skuRepo repository.SKURepository, invSvc StoreInventoryService, db *gorm.DB, logger *slog.Logger) StockRecordService {
	return &stockRecordService{recordRepo: recordRepo, invRepo: invRepo, storeRepo: storeRepo, skuRepo: skuRepo, invSvc: invSvc, db: db, logger: logger}
}

func (s *stockRecordService) Create(storeID, skuID uint, recordType constants.StockRecordType, quantity int, relatedOrderID *uint) (*model.StockRecord, error) {
	var record *model.StockRecord
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var err error
		record, err = s.CreateTx(tx, storeID, skuID, recordType, quantity, relatedOrderID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s *stockRecordService) CreateTx(tx *gorm.DB, storeID, skuID uint, recordType constants.StockRecordType, quantity int, relatedOrderID *uint) (*model.StockRecord, error) {
	if !recordType.Valid() || quantity <= 0 {
		return nil, fmt.Errorf("create stock record type[%s] qty[%d]: %w", recordType, quantity, util.ErrValidation)
	}
	delta := constants.StockDirection(recordType) * quantity
	if err := s.invSvc.AdjustQuantityTx(tx, storeID, skuID, delta); err != nil {
		return nil, fmt.Errorf("create stock record store[%d] sku[%d]: %w", storeID, skuID, err)
	}
	record := &model.StockRecord{
		StoreID: storeID, SKUID: skuID, RecordType: recordType,
		Quantity: quantity, RelatedOrderID: relatedOrderID,
	}
	if err := s.recordRepo.CreateTx(tx, record); err != nil {
		return nil, fmt.Errorf("create stock record: %w", err)
	}
	s.logger.Info(constants.LogStockRecordCreated, "record_id", record.ID, "store", storeID, "sku", skuID, "type", recordType, "qty", quantity)
	return record, nil
}

func (s *stockRecordService) List(page, pageSize int, storeID, skuID uint, recordType constants.StockRecordType) ([]model.StockRecord, int64, error) {
	records, total, err := s.recordRepo.List(page, pageSize, storeID, skuID, recordType)
	if err != nil {
		return nil, 0, fmt.Errorf("list stock records: %w", err)
	}
	s.logger.Info(constants.LogStockRecordListQueried, "total", total)
	return records, total, nil
}

func (s *stockRecordService) Export(storeID uint) ([]model.StockRecord, error) {
	records, _, err := s.recordRepo.List(1, 10000, storeID, 0, "")
	if err != nil {
		return nil, fmt.Errorf("export stock records: %w", err)
	}
	s.logger.Info(constants.LogStockRecordExported, "count", len(records))
	return records, nil
}

func (s *stockRecordService) CreateStocktake(storeID, skuID uint, stocktakeDate string, actualQty int, remark string) (*model.Stocktake, error) {
	if actualQty < 0 {
		return nil, fmt.Errorf("create stocktake actual qty[%d]: %w", actualQty, util.ErrValidation)
	}
	var st *model.Stocktake
	err := s.db.Transaction(func(tx *gorm.DB) error {
		inv, err := s.invRepo.FindByStoreAndSKUTx(tx, storeID, skuID)
		if err != nil {
			return fmt.Errorf("create stocktake store[%d] sku[%d]: %w", storeID, skuID, err)
		}
		difference := actualQty - inv.Quantity
		st = &model.Stocktake{
			StoreID: storeID, SKUID: skuID, StocktakeDate: stocktakeDate,
			SystemQty: inv.Quantity, ActualQty: actualQty, Difference: difference, Remark: remark,
		}
		if err := s.recordRepo.CreateStocktakeTx(tx, st); err != nil {
			return fmt.Errorf("create stocktake: %w", err)
		}
		if difference != 0 {
			if err := s.invRepo.AdjustQuantityTx(tx, storeID, skuID, difference); err != nil {
				return fmt.Errorf("create stocktake adjust inventory: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogStocktakeCreated, "stocktake_id", st.ID, "difference", st.Difference)
	return st, nil
}

func (s *stockRecordService) ListStocktakes(page, pageSize int, storeID uint) ([]model.Stocktake, int64, error) {
	list, total, err := s.recordRepo.ListStocktakes(page, pageSize, storeID)
	if err != nil {
		return nil, 0, fmt.Errorf("list stocktakes: %w", err)
	}
	s.logger.Info(constants.LogStocktakeListQueried, "total", total)
	return list, total, nil
}

func (s *stockRecordService) ReplenishSuggestions() ([]util.ReplenishSuggestion, error) {
	alerts, err := s.invRepo.ListAlerts()
	if err != nil {
		return nil, fmt.Errorf("replenish suggestions: %w", err)
	}
	stores, err := s.storeRepo.ListAll()
	if err != nil {
		return nil, fmt.Errorf("replenish suggestions stores: %w", err)
	}
	storeName := make(map[uint]string)
	for _, st := range stores {
		storeName[st.ID] = st.Name
	}
	var suggestions []util.ReplenishSuggestion
	for _, inv := range alerts {
		sku, err := s.skuRepo.FindByID(inv.SKUID)
		if err != nil {
			continue
		}
		monthlySales, _ := s.recordRepo.MonthlySalesBySKU(inv.StoreID, inv.SKUID)
		suggestions = append(suggestions, util.ReplenishSuggestion{
			SkuID: sku.ID, SkuCode: sku.Code, SkuName: sku.Name,
			StoreID: inv.StoreID, StoreName: storeName[inv.StoreID],
			Quantity: inv.Quantity, SafetyStock: inv.SafetyStock,
			SuggestQty: util.CalculateSuggestQty(inv.Quantity, inv.SafetyStock),
			SlowMoving: util.IsSlowMoving(inv.Quantity, monthlySales),
			Reason:     "低库存预警，建议补货至安全库存的1.5倍",
		})
	}
	s.logger.Info(constants.LogReplenishSuggested, "count", len(suggestions))
	return suggestions, nil
}
