package handler_test

import (
	"github.com/Jaruvat303/cashlog/internal/delivery/http/middleware"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// newTestApp สร้าง fiber app พร้อม ErrorHandler ตัวจริง (ตัวเดียวกับที่ใช้ใน main.go)
// เพื่อให้ error mapping เช่น domain.ErrNotFound -> 404 ทำงานเหมือนตอน production จริงๆ
// แทนที่จะพึ่ง default error handler ของ fiber ซึ่งจะ leak err.Error() ดิบๆ ออกมาแทน
func newTestApp(appLogger logger.Logger) *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: middleware.NewGlobalErrorHandler(appLogger),
	})
}
