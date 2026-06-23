package middleware

import (
	"context"

	"github.com/Jaruvat303/cashlog/pkg/timeutil"
	"github.com/gofiber/fiber/v2"
)

func NewTimezoneMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// นำพิกัดเวลาไทยฝังลงไปใน context ของ Request นั้น ๆ
		ctx := context.WithValue(c.UserContext(), "app_location", timeutil.BangKokLoc)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
