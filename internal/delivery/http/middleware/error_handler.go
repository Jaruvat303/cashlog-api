package middleware

import (
	"errors"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
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

		// เป็น Error ที่มาจากตัว Framework (Fiber Error)
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
			switch {
			// 400 Bad Request
			case errors.Is(err, domain.ErrInvalidInput):
				statusCode = fiber.StatusBadRequest
				errorCode = "INVALID_INPUT_PARAMETERS"
				clientMessage = err.Error()

			// 404 Not Found
			case errors.Is(err, domain.ErrNotFound):
				statusCode = fiber.StatusNotFound
				errorCode = "RESOURCE_NOT_FOUND"
				clientMessage = "The requested data was not found"

			// 409 Conflict
			case errors.Is(err, domain.ErrDuplicateRequest):
				statusCode = fiber.StatusConflict
				errorCode = "DUPLICATE_RESOURCE" // 💡 ช่วยแก้คำสะกดผิดจาก RESOURSE ให้เป็น RESOURCE ครับ
				clientMessage = "This data already exists in our system."

			// 499 Client Closed Request (มาตรฐานสากลสำหรับเคส Client กดยกเลิกกลางคัน)
			case errors.Is(err, domain.ErrContextCanceled):
				statusCode = 499
				errorCode = "REQUEST_CANCELED"
				clientMessage = "The request was canceled by the user."

			// 504 Gateway Timeout (ฐานข้อมูลทำงานนานเกินเวลากำหนด)
			case errors.Is(err, domain.ErrTimeout):
				statusCode = fiber.StatusGatewayTimeout
				errorCode = "DATABASE_TIMEOUT"
				clientMessage = "The database operation timed out, please try again."

			// 500 Internal Server Error จากฐานข้อมูล
			case errors.Is(err, domain.ErrInternalDB):
				statusCode = fiber.StatusInternalServerError
				errorCode = "INTERNAL_DATABASE_ERROR"
				clientMessage = "Something went wrong, please try again later."
			}
		}

		// พ่น Log แยกตามระดับความรุนแรง (เปลี่ยนเป็นเช็กตระกูล >= 500)
		if statusCode >= fiber.StatusInternalServerError {
			// พังที่ระบบหลังบ้านหรือระบบล่ม -> ระดับ Error (ต้องรีบเข้าแก้ไข)
			reqLogger.Error("server-side error caught by global handler",
				zap.Error(err),
				zap.String("error_code", errorCode),
			)
		} else {
			// Client ทำพังเอง หรือกดยกเลิกเอง -> ระดับ Warn
			reqLogger.Warn("client-side side-effect caught by global handler",
				zap.Error(err),
				zap.Int("http_status", statusCode),
				zap.String("error_code", errorCode),
			)
		}

		// ส่ง JSON มาตรฐานกลับไปหาหน้าบ้าน
		return c.Status(statusCode).JSON(dto.ErrorResponseDTO{
			Success:   false,
			ErrorCode: errorCode,
			Message:   clientMessage,
		})
	}
}
