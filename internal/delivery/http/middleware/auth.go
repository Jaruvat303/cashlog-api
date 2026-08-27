package middleware

import (
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(apiKey string, appLogger logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("X-API-Key") != apiKey {
			appLogger.Warn("Unauthorized API request received")
			return fiber.NewError(fiber.StatusUnauthorized, "Missing or invalid API key.")
		}
		return c.Next()
	}
}
