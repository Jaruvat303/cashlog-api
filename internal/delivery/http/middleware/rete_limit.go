package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// NewRateLimiter สร้าง Middleware สำหรับ Rate Limiting โดยใช้ Fiber
func NewRateLimiter(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,        // จำนวนคำขอสูงสุดที่อนุญาตในช่วงเวลาที่กำหนด
		Expiration: expiration, // ระยะเวลาที่กำหนดสำหรับการนับคำขอ
		KeyGenerator: func(c *fiber.Ctx) string {
			clientIP := c.Get("X-Forwarded-For") // ตรวจสอบ Header X-Forwarded-For ก่อน
			if clientIP == "" {
				clientIP = c.IP() // หากไม่มี Header ให้ใช้ IP ของผู้ใช้โดยตรง
			}
			return clientIP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too many requests",
				"message": "You have exceeded the maximum number of requests. Please try again later.",
			})
		},
	})
}
