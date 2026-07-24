package main

import (
	"context"
	"log"

	"github.com/Jaruvat303/cashlog/cmd/config"
	"github.com/Jaruvat303/cashlog/internal/delivery/http/middleware"
	"github.com/Jaruvat303/cashlog/internal/delivery/http/router"
	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/handler"
	geminiClient "github.com/Jaruvat303/cashlog/internal/infrastructure/gemini"
	"github.com/Jaruvat303/cashlog/internal/repository/postgres"
	"github.com/Jaruvat303/cashlog/internal/repository/redis"
	"github.com/Jaruvat303/cashlog/internal/usecase"
	"github.com/Jaruvat303/cashlog/pkg/database"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// @title CashLog API
// @version 1.0
// @description API บริการหลังบ้านสำหรับจัดการบันทึกรายรับรายจ่ายและวิเคราะห์สลิปด้วย Gemini AI
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email jaruvat.see@gmail.com

// @BasePath /api/v1

func main() {
	// Load Config
	cfg := config.LoadConfig()

	// เรียกใช้งาน Zap Logger
	appLogger := logger.InitLogger(cfg.AppEnv)

	//  สร้างรูทคลาวด์ Context และฉีด Logger เข้าไปเป็นศูนย์กลางของระบบ
	ctx := context.Background()
	ctx = logger.WithContext(ctx, appLogger)

	// Connect database
	db := database.InitPostgresDB(ctx, cfg)
	rdb := database.InitRedisDB(ctx, cfg)

	// Create Seed Category Data
	err := database.SeedCategories(ctx, db)
	if err != nil {
		appLogger.Fatal("Warning: failed to seed categories: ", zap.Error(err))
	}

	// Dependency Injection
	txRepository := postgres.NewGormTransactionRepository(db, appLogger)
	cacheRepository := redis.NewRedisDashboardRepository(rdb)
	geminiClient, err := geminiClient.NewClient(ctx, *cfg)
	if err != nil {
		appLogger.Fatal("Warning: failed to create geminiClient: ", zap.Error(err))
	}

	categoryRepository := postgres.NewGORMCategoryRepository(db, appLogger)

	// Inject เลเยอร์นอกเข้าไปใน Layer Usecase
	txUsecase := usecase.NewTransactionUsecase(txRepository, cacheRepository, geminiClient, appLogger)
	catUsecase := usecase.NewCategoryUsecase(categoryRepository, appLogger)

	// Inject Usecase เข้าไปใน Handler
	txhandler := handler.NewTransactionHandler(txUsecase, appLogger)
	catHandler := handler.NewCategoryHandler(catUsecase, appLogger)

	// Health Check Handler
	healthHandler := handler.NewHealthHandler(db, rdb)

	app := fiber.New(fiber.Config{
		AppName:      "CashLog API v1.0",
		ErrorHandler: middleware.NewGlobalErrorHandler(appLogger),
	})

	// Middleware Setting
	app.Use(middleware.NewRecoverMiddleware(appLogger))  // ต้องอยู่บนสุดเพื่อดักจับจุดตาย
	app.Use(middleware.NewCORSMiddleware())              // อนุญาตให้หน้าบ้าน Flutter ยิงเข้ามาได้
	app.Use(middleware.NewRequestLogger(cfg, appLogger)) // เปิดระบบ Structured JSON Log บันทึกลง Cloud
	app.Use(middleware.NewTimezoneMiddleware())          // ล็อกเวลาสากลในระบบให้เป็นเวลาไทยเสมอ

	// ส่ง handler เพื่อสร้าง Http route
	router.SetupRoutes(app, txhandler, catHandler, healthHandler)

	// 6. สั่งเปิดเซิร์ฟเวอร์รันระบบตามพอร์ตที่กำหนด
	log.Printf("🚀 CashLog API runs smoothly on environment [%s]", cfg.AppEnv)
	log.Fatal(app.Listen("0.0.0.0:" + cfg.Port))
}
