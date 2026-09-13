package router

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/ld/storeinventory/internal/config"
	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/handler"
	"github.com/ld/storeinventory/internal/middleware"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/repository"
	"github.com/ld/storeinventory/internal/service"
)

// Router 装配依赖并注册路由。
type Router struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *gorm.DB
}

// NewRouter 初始化数据库、依赖并返回 gin 引擎。
func NewRouter(cfg *config.Config, logger *slog.Logger) (*gin.Engine, error) {
	r := &Router{cfg: cfg, logger: logger}
	if err := r.connectDB(); err != nil {
		return nil, err
	}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	if err := r.seed(); err != nil {
		return nil, err
	}
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.ErrorHandler())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     r.cfg.CORSOriginsSlice(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	engine.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	api := engine.Group("/api")
	api.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.registerV1(api.Group("/v1"))
	return engine, nil
}

func (r *Router) connectDB() error {
	var db *gorm.DB
	var err error
	for i := 0; i < 30; i++ {
		db, err = gorm.Open(postgres.Open(r.cfg.DSN()), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr == nil {
				sqlDB.SetMaxOpenConns(20)
				sqlDB.SetMaxIdleConns(5)
			}
			r.db = db
			r.logger.Info("database connected", "host", r.cfg.DBHost, "db", r.cfg.DBName)
			return nil
		}
		r.logger.Warn("database not ready, retrying", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("connect database: %w", err)
}

func (r *Router) migrate() error {
	if err := r.db.AutoMigrate(
		&model.User{}, &model.Store{}, &model.SKU{}, &model.StoreInventory{},
		&model.TransferOrder{}, &model.StockRecord{}, &model.Stocktake{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	r.logger.Info("database migrated")
	return nil
}

// seed 初始化种子数据（仅当用户表为空时执行，database/init.sql 已包含同源数据）。
func (r *Router) seed() error {
	var count int64
	if err := r.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count users for seed: %w", err)
	}
	if count > 0 {
		r.logger.Info("seed skipped, users exist")
		return nil
	}
	if err := seedData(r.db, r.logger); err != nil {
		return fmt.Errorf("seed data: %w", err)
	}
	r.logger.Info("seed data inserted")
	return nil
}

func (r *Router) registerV1(v1 *gin.RouterGroup) {
	userRepo := repository.NewUserRepository(r.db)
	storeRepo := repository.NewStoreRepository(r.db)
	skuRepo := repository.NewSKURepository(r.db)
	invRepo := repository.NewStoreInventoryRepository(r.db)
	transferRepo := repository.NewTransferOrderRepository(r.db)
	recordRepo := repository.NewStockRecordRepository(r.db)

	userSvc := service.NewUserService(userRepo, r.logger, r.cfg.JWTSecret, r.cfg.TokenTTLHours)
	storeSvc := service.NewStoreService(storeRepo, r.logger)
	skuSvc := service.NewSKUService(skuRepo, r.logger)
	invSvc := service.NewStoreInventoryService(invRepo, skuRepo, r.db, r.logger)
	recordSvc := service.NewStockRecordService(recordRepo, invRepo, storeRepo, skuRepo, invSvc, r.db, r.logger)
	transferSvc := service.NewTransferOrderService(transferRepo, invSvc, recordSvc, r.db, r.logger)

	userHandler := handler.NewUserHandler(userSvc)
	storeHandler := handler.NewStoreHandler(storeSvc)
	skuHandler := handler.NewSKUHandler(skuSvc)
	invHandler := handler.NewStoreInventoryHandler(invSvc)
	transferHandler := handler.NewTransferOrderHandler(transferSvc)
	recordHandler := handler.NewStockRecordHandler(recordSvc)

	auth := middleware.AuthRequired(r.cfg)
	authLimiter := middleware.RateLimitStrict(r.cfg)
	adminRoles := []constants.UserRole{constants.RoleAdmin, constants.RoleHQ}
	managerRoles := []constants.UserRole{constants.RoleAdmin, constants.RoleHQ, constants.RoleStoreManager}

	registerAuthRoutes(v1, userHandler, authLimiter)
	registerUserRoutes(v1, userHandler, auth)
	registerStoreRoutes(v1, storeHandler, auth, adminRoles)
	registerSKURoutes(v1, skuHandler, auth, adminRoles, authLimiter)
	registerInventoryRoutes(v1, invHandler, auth, managerRoles)
	registerTransferRoutes(v1, transferHandler, auth, managerRoles, authLimiter)
	registerStockRecordRoutes(v1, recordHandler, auth, managerRoles)
}
