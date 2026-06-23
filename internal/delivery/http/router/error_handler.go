package router

import (
	"errors"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// GlobalErrorHandler
func NewGlobalErrorHandler(appLogger logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		ctx := c.UserContext()

		// ดึง Logger จาก Context ถ้าไม่มีจะได้ค่า nil กลับมา
		reqLogger := logger.Ctx(ctx)

		// ถ้าเป็น nil ให้ใช้ appLogger (Global ตัวหลัก) ทันที
		if reqLogger == nil {
			reqLogger = appLogger
		}

		// กำหนดค่ามารตฐานสำหรับ Error ทั่วไป (500 Internal Server Error)
		statusCode := fiber.StatusInternalServerError
		errorCode := "INTERNAL_SERVER_ERROR"
		clientMessage := "Something went wrong, please try again later."

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			statusCode = fiberErr.Code
			clientMessage = fiberErr.Message

			// ลิงค์ Error Code ให้สอดคล่องกับ HTTP Status ตัวนั้นๆ
			switch fiberErr.Code {
			case fiber.StatusBadRequest:
				errorCode = "BAD_REQUEST_PARAMETERS"
			case fiber.StatusNotFound:
				errorCode = "URL_NOT_FOUND"
			default:
				errorCode = "CLIENT_ERROR"
			}
		} else {
			// ใช้ Error.Is เช็กประเภท เพื่อแมป HTTP Status ให้ถูกต้อง
			if errors.Is(err, domain.ErrInvalidInput) {
				statusCode = fiber.StatusBadRequest
				errorCode = "INVALID_INPUT_PARAMETERS"
				clientMessage = err.Error()
			} else if errors.Is(err, domain.ErrNotFound) {
				statusCode = fiber.StatusNotFound
				errorCode = "RESOURCE_NOT_FOUND"
				clientMessage = "The requested data was not found"
			} else if errors.Is(err, domain.ErrDuplicateRequest) {
				statusCode = fiber.StatusConflict
				errorCode = "DUPLICATE_RESOURSE"
				clientMessage = "This data already exists in our system."
			}
		}

		// พ่น Log ผ่าน Zap ตามระดับความรุนแรง
		if statusCode == fiber.StatusInternalServerError {
			// พังที่ระบบเราเอง -> ระดับ Error
			reqLogger.Error("internal server error caught by global handler", zap.Error(err))
		} else {
			// client ทำพังเอง -> ระดับ Warn
			reqLogger.Warn("client error caught by global handler",
				zap.Error(err),
				zap.Int("http_status", statusCode),
			)
		}

		// ส่ง JSON มาตรฐานกลับไปหาหน้าบ้าน
		return c.Status(statusCode).JSON(fiber.Map{
			"success":    false,
			"error_code": errorCode,
			"message":    clientMessage,
		})
	}
}
