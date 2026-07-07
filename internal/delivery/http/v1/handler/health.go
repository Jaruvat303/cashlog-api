package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := "UP"
	services := fiber.Map{}

	// 1. Check Supabase (PostgresSQL via GORM)
	sqlDB, err := h.db.DB()
	if err != nil {
		status = "DOWN"
		services["database"] = "DOWN: " + err.Error()
	} else {
		if err := sqlDB.PingContext(ctx); err != nil {
			status = "DOWN"
			services["database"] = "DOWN: " + err.Error()
		} else {
			services["database"] = "OK"
		}
	}

	// 2. Check Upstash (Redis)
	if err := h.redis.Ping(ctx).Err(); err != nil {
		status = "DOWN"
		services["redis"] = "DOWN: " + err.Error()
	} else {
		services["redis"] = "OK"
	}

	// 3. Check Gemini API Connection (Optional - Check via simple client init if needed)
	// เนื่องจาก Gemini เป็น External API มักจะพึ่งพาการยิง HTTP จึงไม่นิยมทำ Ping ตรงๆ ใน Health Check
	// เพื่อป้องกันไม่ให้ระบบเรา DOWN เวลาที่ API ภายนอกล่มชั่วคราว แต่ระบุสถานะไว้ได้
	services["gemini_api"] = "CONFIGURED"

	// ถ้าบริการหลักพัง ส่ง HTTP 503 Services Unavailable กลับไปเพื่อให้ Cloud Run หรือ Load Balancer รู้
	if status == "DOWN" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":    status,
			"servicess": services,
			"time":      time.Now().Format(time.RFC3339),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    status,
		"servicess": services,
		"time":      time.Now().Format(time.RFC3339),
	})
}
