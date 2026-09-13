-- cy-364 连锁门店库存调配系统 初始化脚本（PostgreSQL 15）
-- 首次启动容器时由 /docker-entrypoint-initdb.d 自动执行，全部幂等。

CREATE TABLE IF NOT EXISTS stores (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL CONSTRAINT uni_stores_code UNIQUE,
    name VARCHAR(128) NOT NULL,
    address VARCHAR(255) DEFAULT '',
    manager_user_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) NOT NULL CONSTRAINT uni_users_username UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(64) DEFAULT '',
    role VARCHAR(32) NOT NULL DEFAULT 'store_manager',
    store_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_users_store_id ON users(store_id);

CREATE TABLE IF NOT EXISTS skus (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL CONSTRAINT uni_skus_code UNIQUE,
    name VARCHAR(128) NOT NULL,
    spec VARCHAR(128) DEFAULT '',
    barcode VARCHAR(64),
    category VARCHAR(64),
    unit VARCHAR(16),
    status VARCHAR(16) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_skus_category ON skus(category);
CREATE INDEX IF NOT EXISTS idx_skus_barcode ON skus(barcode);

CREATE TABLE IF NOT EXISTS store_inventories (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    safety_stock INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_store_inventories_store ON store_inventories(store_id);
CREATE INDEX IF NOT EXISTS idx_store_inventories_sku ON store_inventories(sku_id);

CREATE TABLE IF NOT EXISTS transfer_orders (
    id BIGSERIAL PRIMARY KEY,
    from_store_id BIGINT NOT NULL,
    to_store_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    reason VARCHAR(255) DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_transfer_orders_from ON transfer_orders(from_store_id);
CREATE INDEX IF NOT EXISTS idx_transfer_orders_to ON transfer_orders(to_store_id);
CREATE INDEX IF NOT EXISTS idx_transfer_orders_status ON transfer_orders(status);

CREATE TABLE IF NOT EXISTS stock_records (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL,
    record_type VARCHAR(16) NOT NULL,
    quantity INT NOT NULL,
    related_order_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_stock_records_store ON stock_records(store_id);
CREATE INDEX IF NOT EXISTS idx_stock_records_sku ON stock_records(sku_id);
CREATE INDEX IF NOT EXISTS idx_stock_records_type ON stock_records(record_type);

CREATE TABLE IF NOT EXISTS stocktakes (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL,
    stocktake_date VARCHAR(16),
    system_qty INT NOT NULL DEFAULT 0,
    actual_qty INT NOT NULL DEFAULT 0,
    difference INT NOT NULL DEFAULT 0,
    remark VARCHAR(255) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_stocktakes_store ON stocktakes(store_id);

-- ============ 种子数据 ============
INSERT INTO stores (id, code, name, address) VALUES
    (1, 'ST001', '北京朝阳门店', '北京市朝阳区建国路 88 号'),
    (2, 'ST002', '上海静安门店', '上海市静安区南京西路 1266 号'),
    (3, 'ST003', '广州天河门店', '广州市天河区天河路 208 号')
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, username, password_hash, name, role, store_id) VALUES
    (1, 'admin', '$2b$10$6unwpWA5bcj7.bqT2KaI/uTvGhlwDEUHhIGlRbKFLW149cmOKjyNK', '系统管理员', 'admin', NULL),
    (2, 'hquser', '$2b$10$N9syUa1g6oc9kJeDBX3y2.knOgByWgtWOcpB8OXvO5sR6kyy.HfPO', '总部采购', 'hq', NULL),
    (3, 'manager1', '$2b$10$/29y.6Y3uosx33E0ZpDRiOGEKAoUX3ULcZ4KPMssUb5/CnBBr.utK', '朝阳店长', 'store_manager', 1),
    (4, 'manager2', '$2b$10$/29y.6Y3uosx33E0ZpDRiOGEKAoUX3ULcZ4KPMssUb5/CnBBr.utK', '静安店长', 'store_manager', 2)
ON CONFLICT (id) DO NOTHING;

INSERT INTO skus (id, code, name, spec, barcode, category, unit, status) VALUES
    (1, 'SKU1001', '可口可乐 330ml', '330ml/罐', '6902538001000', '饮料', '罐', 'active'),
    (2, 'SKU1002', '农夫山泉 550ml', '550ml/瓶', '6902538002000', '饮料', '瓶', 'active'),
    (3, 'SKU2001', '康师傅红烧牛肉面', '袋装 105g', '6902538003000', '方便食品', '袋', 'active'),
    (4, 'SKU2002', '乐事薯片原味', '70g/袋', '6902538004000', '零食', '袋', 'active'),
    (5, 'SKU3001', '维达抽纸 3 层', '120 抽/包', '6902538005000', '日用品', '包', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO store_inventories (id, store_id, sku_id, quantity, safety_stock) VALUES
    (1, 1, 1, 120, 50),
    (2, 1, 2, 20, 80),
    (3, 1, 3, 200, 60),
    (4, 2, 1, 30, 40),
    (5, 2, 4, 150, 50),
    (6, 3, 5, 8, 30)
ON CONFLICT (id) DO NOTHING;

INSERT INTO stock_records (id, store_id, sku_id, record_type, quantity) VALUES
    (1, 1, 1, 'purchase', 200),
    (2, 2, 1, 'purchase', 150),
    (3, 1, 2, 'sale', 30)
ON CONFLICT (id) DO NOTHING;

SELECT setval('users_id_seq', GREATEST((SELECT MAX(id) FROM users), 1));
SELECT setval('stores_id_seq', GREATEST((SELECT MAX(id) FROM stores), 1));
SELECT setval('skus_id_seq', GREATEST((SELECT MAX(id) FROM skus), 1));
SELECT setval('store_inventories_id_seq', GREATEST((SELECT MAX(id) FROM store_inventories), 1));
SELECT setval('stock_records_id_seq', GREATEST((SELECT MAX(id) FROM stock_records), 1));
