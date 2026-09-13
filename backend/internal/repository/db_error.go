package repository

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgconn"
)

// dbOrTx 返回事务连接（若 tx 非空）或默认连接，用于让仓储方法在事务内外复用。
func dbOrTx(db, tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return db
}

// isDuplicatePostgres 识别 PostgreSQL 唯一键冲突。
func isDuplicatePostgres(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// isDuplicateMySQL 识别 MySQL 唯一键冲突（兼容未来扩展）。
func isDuplicateMySQL(err error) bool {
	return false
}

// isDuplicate 兼容 PostgreSQL / MySQL / SQLite 错误文本。
func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	if isDuplicatePostgres(err) || isDuplicateMySQL(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
