package database

import (
	"context"
	_ "embed"
	"time"

	"github.com/Jaruvat303/cashlog/cmd/config"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// InitPostgresDB ทำหน้าที่เปิดการเชื่อมต่อฐานข้อมูลตาม Config และส่งกลับ instance ของ *gorm.DB
func InitPostgresDB(ctx context.Context, cfg *config.Config) *gorm.DB {
	if cfg.DBURL == "" {
		logger.Ctx(ctx).Fatal("Configuration 'DB_URL' is required but found empty")
	}

	// ตั้งค่าการแสดง Log SQL
	var gormLogMode gormLogger.Interface
	if cfg.AppEnv != "production" {
		gormLogMode = gormLogger.Default.LogMode(gormLogger.Info)
	} else {
		gormLogMode = gormLogger.Default.LogMode(gormLogger.Error)
	}

	// เริ่มเปิดการเชื่อมต่อ
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DBURL,
		PreferSimpleProtocol: true, //เพื่อป้องกัน Error จากการทำ Prepare Statement ซ้ำซ้อน
	}), &gorm.Config{
		Logger: gormLogMode,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		logger.Ctx(ctx).Fatal("Failed to connect to PostgresSQL: %v", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Ctx(ctx).Fatal("Failed to retrieve generic SQL instance: %v", zap.Error(err))
	}

	// ตั้งค่า Connection Pool โดยหยิบค่ามาจากสัญญลักษณ์โตรงสร้างโดยตรง
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		logger.Ctx(ctx).Fatal("PostgresSQL ping failed: %v", zap.Error(err))
	}

	// สั่งให้ GORM ตรวจสอบและสร้างตาราง "transactions" บน Supabase อัตโนมัติ
	logger.Ctx(ctx).Info("⏳ Running Database Auto Migration...")
	err = db.AutoMigrate(
		&domain.Category{},
		&domain.Transaction{},
	) // ปรับให้ตรงกับชื่อ Struct ตาราง
	if err != nil {
		logger.Ctx(ctx).Fatal("❌ Database Migration Failed: %v", zap.Error(err))
	}

	logger.Ctx(ctx).Info("✅ Database Migration Completed Successfully!")
	logger.Ctx(ctx).Info("PostgresSQL database connection established cleanly via Config Struct")

	return db
}

//go:embed init_categories.sql
var initCategoriesSQL string

// SeedCategories ทำหน้าที่อ่านไฟล์ SQL และ Execute เข้า Database
func SeedCategories(ctx context.Context, db *gorm.DB) error {

	// รันคำสั่ง SQL ผ่าน GORM
	if err := db.Exec(initCategoriesSQL).Error; err != nil {
		return err
	}

	logger.Ctx(ctx).Info("Successfully seeded categories data!")
	return nil

}
