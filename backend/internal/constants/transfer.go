package constants

// TransferStatus 调拨单状态枚举，前后端共享定义（frontend/src/constants/transfer.ts 对应实现）。
type TransferStatus string

const (
	TransferPending   TransferStatus = "pending"
	TransferConfirmed TransferStatus = "confirmed"
	TransferShipped   TransferStatus = "shipped"
	TransferReceived  TransferStatus = "received"
	TransferCancelled TransferStatus = "cancelled"
)

func (s TransferStatus) Valid() bool {
	switch s {
	case TransferPending, TransferConfirmed, TransferShipped, TransferReceived, TransferCancelled:
		return true
	}
	return false
}

// TransferStatusFlow 定义调拨单允许的合法状态流转，状态机在 service、前端按钮、日志模板、错误码、formatters 中多处耦合。
var TransferStatusFlow = map[TransferStatus][]TransferStatus{
	TransferPending:   {TransferConfirmed, TransferCancelled},
	TransferConfirmed: {TransferShipped, TransferCancelled},
	TransferShipped:   {TransferReceived, TransferCancelled},
	TransferReceived:  {},
	TransferCancelled: {},
}

func CanTransfer(from, to TransferStatus) bool {
	for _, next := range TransferStatusFlow[from] {
		if next == to {
			return true
		}
	}
	return false
}
