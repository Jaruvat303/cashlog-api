package middleware

import (
	"context"

	"github.com/Jaruvat303/cashlog/pkg/timeutil"
	"github.com/gofiber/fiber/v2"
)

type ctxKey string

const appLocationKey ctxKey = "app_location"

func NewTimezoneMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// นำพิกัดเวลาไทยฝังลงไปใน context ของ Request นั้น ๆ
		ctx := context.WithValue(c.UserContext(), appLocationKey, timeutil.BangKokLoc)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
