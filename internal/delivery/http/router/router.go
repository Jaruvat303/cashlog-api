package router

import (
	"time"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/middleware"
	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/swagger"

	_ "github.com/Jaruvat303/cashlog/docs"
)

func SetupRoutes(app *fiber.App,
	txHandler *handler.TransactionHandler,
	catHandler *handler.CategoryHandler,
	accHandler *handler.AccountHandler,
	healthHandler *handler.HealthHandler,
) {

	// ==========================================
	// System & DevOps Routes (วางไว้บนสุด ไม่ผ่าน Middleware คุมสิทธิ์)
	// ==========================================

	// Health Check
	app.Get("/health", healthHandler.Check)

	// Metrics สำหรับ Prometheus/Grafana หรือเปิดดูผ่าน Browser เองได้
	app.Get("/metrics", monitor.New(monitor.Config{
		Title: "My API Performance Metrics",
	}))

	// Swagger API Documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Rate Limiting Middleware
	generalRateLimiter := middleware.NewRateLimiter(60, 60*time.Second) // จำกัดคำขอทั่วไป 60 ครั้งต่อ 60 วินาที
	slipRateLimiter := middleware.NewRateLimiter(10, 60*time.Second)    // จำกัดคำขอสำหรับการอัปโหลดสลิป 10 ครั้งต่อ 60 วินาที

	// กำหนดและผูกเส้นทาง API Endpoint ตามสัญญา RESTful Spec
	v1 := app.Group("/api/v1", generalRateLimiter) // ใช้ Rate Limiter สำหรับทุกคำขอในกลุ่มนี้

	tx := v1.Group("/transactions")
	tx.Get("/", txHandler.GetMonthlyHistory)
	tx.Get("/summary", txHandler.GetDashboardSummary)
	tx.Post("/upload-slip", slipRateLimiter, txHandler.UplaodSlipAndLog)
	tx.Patch("/:id", txHandler.UpdateTransaction)
	tx.Delete("/:id", txHandler.DeleteTransaction)

	cat := v1.Group("/categories")
	cat.Post("/", catHandler.CreateCategory)
	cat.Get("/", catHandler.FetchCategoriesByType)
	cat.Patch("/:id", catHandler.UpdateCategory)
	cat.Delete("/:id", catHandler.DeleteCategory)

	acc := v1.Group("/accounts")
	acc.Post("/", accHandler.CreateAccount)
	acc.Get("/", accHandler.FetchActiveAccounts)
	acc.Patch("/:id", accHandler.UpdateAccount)
	acc.Delete("/:id", accHandler.DeleteAccount)

}
