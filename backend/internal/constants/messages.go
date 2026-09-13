package constants

// messages.go 同时承载前端提示文案、后端返回文案与日志文案（屎山约束：修改一个文案可能同时影响前端显示、后端返回和日志内容）。
const (
	MsgOK                     = "ok"
	MsgInvalidRequest         = "请求参数不合法"
	MsgUnauthorized           = "未登录或登录已过期"
	MsgForbidden              = "无权限执行该操作"
	MsgNotFound               = "资源不存在"
	MsgConflict               = "资源状态冲突"
	MsgRateLimited            = "请求过于频繁，请稍后再试"
	MsgInternalError          = "服务器内部错误"
	MsgLoginSuccess           = "登录成功"
	MsgRegisterSuccess        = "注册成功"
	MsgPasswordIncorrect      = "用户名或密码错误"
	MsgSkuCodeExists          = "SKU 编码已存在"
	MsgStoreCodeExists        = "门店编码已存在"
	MsgTransferConflict       = "调拨单当前状态不允许该操作"
	MsgStockNotEnough         = "库存不足，无法完成出库"
	MsgTransferCreateSuccess  = "调拨申请已提交"
	MsgTransferConfirmSuccess = "调拨单已确认"
	MsgTransferShipSuccess    = "调拨单已发货"
	MsgTransferReceiveSuccess = "调拨单已收货"
	MsgTransferCancelSuccess  = "调拨单已取消"
	MsgRecordCreatedSuccess   = "出入库记录已创建"
	MsgStocktakeSuccess       = "盘点完成，盘盈盘亏已计算"
	MsgSafetyStockUpdated     = "安全库存已更新"
	MsgInventoryLowWarning    = "存在低库存预警"
)
