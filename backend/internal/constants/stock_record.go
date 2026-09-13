package constants

// StockRecordType 出入库记录类型枚举，前后端共享定义（frontend/src/constants/stockRecord.ts 对应实现）。
type StockRecordType string

const (
	RecordPurchase   StockRecordType = "purchase"
	RecordTransferIn StockRecordType = "transfer_in"
	RecordSale       StockRecordType = "sale"
	RecordLoss       StockRecordType = "loss"
)

func (t StockRecordType) Valid() bool {
	switch t {
	case RecordPurchase, RecordTransferIn, RecordSale, RecordLoss:
		return true
	}
	return false
}

// StockDirection 出入方向：入库为 +，出库为 -，用于库存变动计算。
func StockDirection(t StockRecordType) int {
	switch t {
	case RecordPurchase, RecordTransferIn:
		return 1
	case RecordSale, RecordLoss:
		return -1
	}
	return 0
}
