package middleware

import (
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(apiKey string, appLogger logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ถ้า API_KEY ไม่ได้ตั้งค่าไว้ (deploy ลืมตั้ง env) ต้อง reject ทุก request แทนที่จะเปิดโหว่เงียบๆ
		// (ถ้าไม่มี guard นี้ apiKey=="" จะทำให้ "" != apiKey เป็น false เมื่อ client ไม่ส่ง header มาเลย = auth bypass)
		if apiKey == "" {
			appLogger.Error("API_KEY is not configured; refusing all requests")
			return fiber.NewError(fiber.StatusInternalServerError, "Server misconfiguration.")
		}
		if c.Get("X-API-Key") != apiKey {
			appLogger.Warn("Unauthorized API request received")
			return fiber.NewError(fiber.StatusUnauthorized, "Missing or invalid API key.")
		}
		return c.Next()
	}
}
