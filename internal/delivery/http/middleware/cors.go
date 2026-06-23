package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewCORSMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "*", // ในอนาคตถ้าขึ้น Production สามารถใส่ Domain เฉพาะได้
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: false,
	})
}
