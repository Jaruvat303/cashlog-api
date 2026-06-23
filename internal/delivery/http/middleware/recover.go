package middleware

import (
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func NewRecoverMiddleware(appLogger logger.Logger) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		// เขียนตรรกะให้มันพ่นความผิดพลาดผ่าน Zap Logger แทนการ Print ธรรมดา
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			appLogger.Error("Application Panic Recovered!",
				zap.Any("panic_error", e),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)
		},
	})
}
