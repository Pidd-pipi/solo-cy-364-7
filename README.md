# StockLink（连锁门店库存调配系统）

面向连锁零售企业的库存运营平台，覆盖多门店 SKU 统一管理、库存实时同步预警、调拨申请审批、出入库记录盘点、滞销品分析与补货建议等场景。

## Docker Compose 一键启动

```bash
cp .env.example .env
docker compose up -d --build
```

- 前端：http://localhost:28504
- 后端 API：http://localhost:29504（`/api/v1`）
- 数据库：PostgreSQL 15（localhost:5432）

种子账号：

| 账号 | 密码 | 角色 |
| --- | --- | --- |
| admin | admin123 | 管理员 |
| hquser | hq123456 | 总部 |
| manager1 | sm123456 | 店长（北京朝阳门店） |
| manager2 | sm123456 | 店长（上海静安门店） |

## 项目主要功能

- 多门店 SKU 主数据统一维护与批量导入
- 门店库存实时同步、安全库存阈值与低库存预警
- 调拨申请 → 审批确认 → 发货 → 收货全流程状态机
- 出入库明细（采购/调拨/销售/损耗）、周期盘点与盘盈盘亏计算
- 滞销商品分析与智能补货建议报表
- JWT 认证 + RBAC 角色权限（总部/店长/管理员）+ 接口限流

## 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | Vue 3 + TypeScript + Element Plus + Vite |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 15 |
| 认证 | JWT（golang-jwt/v5）+ RBAC |
| 部署 | Docker Compose + Nginx 反向代理 |

## 本地开发

```bash
# 后端
cd backend && go mod tidy && go run ./cmd/server
# 构建
cd backend && go build ./...

# 前端
cd frontend && npm install && npm run dev
```

## 项目目录结构

```
cy-364/
├── backend/
│   ├── cmd/server/main.go
│   └── internal/
│       ├── config/       # 配置解析
│       ├── model/        # 按实体分文件（user/store/sku/store_inventory/transfer_order/stock_record/stocktake）
│       ├── repository/   # 数据访问层
│       ├── service/      # 业务逻辑层
│       ├── handler/      # HTTP 接口层
│       ├── router/       # 路由注册（按实体分文件）
│       ├── middleware/   # auth/rbac/rate_limiter/error_handler/request_logger
│       ├── dto/          # 请求/响应结构体
│       ├── constants/    # 枚举、错误码、日志模板、文案
│       └── util/         # jwt/logger/formatters/app_error/replenish_calculator
├── frontend/
│   └── src/
│       ├── api/          # user/store/sku/storeInventory/transferOrder/stockRecord
│       ├── stores/       # authStore/userStore/inventoryStore/transferStore
│       ├── components/common/  # InventoryStatusBadge/SkuTable/StockLevelIndicator/TransferStatusBadge/RecordTable/ReplenishSuggestionCard/RoleGuard
│       ├── hooks/        # useAuth/useInventoryStats/useTransfers
│       ├── pages/        # Dashboard/Skus/Inventory/Transfers/Records/Analysis/Profile/Login
│       ├── router/       # index.ts + guards.ts
│       ├── utils/        # dateFormat/replenishCalculator/request
│       └── constants/    # transfer/stockRecord/user/errorCodes
├── database/init.sql     # PostgreSQL 初始化脚本（表结构 + 种子数据）
├── docker-compose.yml
├── .env.example
└── README.md
```

## 环境变量

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | Compose 项目名 | ldstoreinventory |
| DB_NAME | 数据库名 | ldstoreinventory_db |
| DB_USER | 数据库用户 | ldstoreinventory_user |
| DB_PASSWORD | 数据库密码 | ldstoreinventory_pwd |
| DB_ROOT_PASSWORD | 兼容模板保留键 | ldstoreinventory_root |
| JWT_SECRET | JWT 签名密钥 | change_me_to_a_long_random_string |
| FRONTEND_PORT | 前端端口 | 28504 |
| BACKEND_PORT | 后端端口 | 29504 |
| DB_PORT | 数据库端口 | 5432 |
| APP_CORS_ORIGINS | 允许跨域来源（逗号分隔） | http://localhost:28504 |

## Docker 部署说明

- 端口映射：前端 `28504:80`、后端 `29504:8080`、数据库 `5432:5432`
- 数据卷：`db_data` 持久化 PostgreSQL 数据
- 服务间通过 Compose 内部网络通信，前端 Nginx 将 `/api/` 反代到 `backend:8080`
- 数据库初始化脚本 `database/init.sql` 在首次启动自动执行（表结构 + 种子数据）
- 常见问题：
  - 端口冲突：修改 `.env` 中的 `FRONTEND_PORT`/`BACKEND_PORT`/`DB_PORT`
  - 数据重置：`docker compose down -v` 后重新 `up -d`
  - 查看日志：`docker compose logs -f backend`

## API 接口清单

| 方法 | 路径 | 说明 | 权限/限流 |
| --- | --- | --- | --- |
| POST | /api/v1/auth/register | 用户注册 | 公开，严格限流 |
| POST | /api/v1/auth/login | 用户登录，返回 JWT | 公开，严格限流 |
| GET | /api/v1/users/me | 查询当前登录用户 | 登录 |
| PUT | /api/v1/users/me | 修改当前用户资料 | 登录 |
| GET | /api/v1/users | 用户分页列表 | 管理员/总部 |
| GET | /api/v1/users/:id | 用户详情 | 管理员/总部 |
| GET | /api/v1/stores | 门店分页列表 | 登录 |
| GET | /api/v1/stores/all | 全量门店列表 | 登录 |
| GET | /api/v1/stores/:id | 门店详情 | 登录 |
| POST | /api/v1/stores | 新增门店 | 管理员/总部 |
| PUT | /api/v1/stores/:id | 修改门店 | 管理员/总部 |
| DELETE | /api/v1/stores/:id | 删除门店 | 管理员/总部 |
| GET | /api/v1/skus | SKU 分页列表 | 登录 |
| POST | /api/v1/skus | 新增 SKU | 管理员/总部 |
| POST | /api/v1/skus/batch-import | 批量导入 SKU | 管理员/总部，严格限流 |
| PUT | /api/v1/skus/:id | 修改 SKU | 管理员/总部 |
| DELETE | /api/v1/skus/:id | 删除 SKU | 管理员/总部 |
| GET | /api/v1/inventories | 门店库存分页列表 | 登录 |
| GET | /api/v1/inventories/stats | 库存总览统计 | 登录 |
| GET | /api/v1/inventories/alerts | 低库存预警列表 | 登录 |
| POST | /api/v1/inventories/ensure | 初始化/确保库存记录 | 店长/管理员/总部 |
| PUT | /api/v1/inventories/:id/safety-stock | 设置安全库存 | 店长/管理员/总部 |
| GET | /api/v1/transfers | 调拨单分页列表 | 登录 |
| POST | /api/v1/transfers | 创建调拨申请 | 店长/管理员/总部，严格限流 |
| PUT | /api/v1/transfers/:id/confirm | 审批确认 | 管理员/总部 |
| PUT | /api/v1/transfers/:id/ship | 发货并扣减调出库存 | 店长/管理员/总部 |
| PUT | /api/v1/transfers/:id/receive | 收货并增加调入库存 | 店长/管理员/总部 |
| PUT | /api/v1/transfers/:id/cancel | 取消调拨单 | 店长/管理员/总部 |
| GET | /api/v1/records | 出入库记录分页列表 | 登录 |
| GET | /api/v1/records/export | 导出出入库记录 | 登录 |
| POST | /api/v1/records | 创建出入库记录并调整库存 | 店长/管理员/总部 |
| POST | /api/v1/records/stocktakes | 创建盘点记录并调整差异 | 店长/管理员/总部 |
| GET | /api/v1/records/stocktakes | 盘点记录分页列表 | 登录 |
| GET | /api/v1/analysis/suggestions | 补货建议报表 | 登录 |
| GET | /healthz | 服务健康检查 | 公开 |
| GET | /api/healthz | API 健康检查 | 公开 |

## 枚举出现位置清单

### TransferStatus（调拨单状态）
- 后端：`backend/internal/constants/transfer.go`（定义 + Valid + TransferStatusFlow + CanTransfer）、`backend/internal/model/transfer_order.go`（GORM 模型）、`backend/internal/service/transfer_order_service.go`（状态机）、`backend/internal/handler/transfer_order_handler.go`（流转接口）、`backend/internal/util/formatters.go`（状态文本）、`backend/internal/constants/log_templates.go`（日志模板）、`backend/internal/constants/error_codes.go`/`messages.go`（文案）
- 前端：`frontend/src/constants/transfer.ts`（定义 + 文案 + 流转）、`frontend/src/types/index.ts`（类型）、`frontend/src/components/common/TransferStatusBadge.vue`（状态徽章）、`frontend/src/pages/Transfers.vue`（按钮显隐与筛选）、`frontend/src/api/transferOrder.ts`

### StockRecordType（出入库类型）
- 后端：`backend/internal/constants/stock_record.go`（定义 + Valid + StockDirection）、`backend/internal/model/stock_record.go`（GORM 模型）、`backend/internal/service/stock_record_service.go`、`backend/internal/handler/stock_record_handler.go`、`backend/internal/util/formatters.go`（类型文本）、`backend/internal/constants/log_templates.go`、`backend/internal/repository/stock_record_repository.go`
- 前端：`frontend/src/constants/stockRecord.ts`（定义 + 文案 + 标签色）、`frontend/src/types/index.ts`、`frontend/src/components/common/RecordTable.vue`（明细展示）、`frontend/src/pages/Records.vue`（筛选与表单）、`frontend/src/api/stockRecord.ts`

### UserRole（用户角色）
- 后端：`backend/internal/constants/user.go`（定义 + Valid）、`backend/internal/model/user.go`（GORM 模型）、`backend/internal/middleware/rbac.go`（RBAC 校验）、`backend/internal/router/*.go`（路由权限）、`backend/internal/service/user_service.go`（注册默认角色）、`backend/internal/util/formatters.go`（角色文本）、`backend/internal/util/jwt.go`（JWT 声明）
- 前端：`frontend/src/constants/user.ts`（定义 + 文案）、`frontend/src/types/index.ts`、`frontend/src/stores/authStore.ts`（角色状态）、`frontend/src/router/guards.ts`（路由守卫）、`frontend/src/components/common/RoleGuard.vue`（按钮/区域显隐）、`frontend/src/pages/Login.vue`、`frontend/src/pages/Layout.vue`（角色标签）

## License

MIT
