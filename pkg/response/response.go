package response

import "github.com/gofiber/fiber/v2"

// JsonResponse เป็นโครงสร้างข้อมูลสำหรับส่งกลับเมื่อทำงานสำเร็จ
type JsonResponse[T any] struct {
	Success bool   `json:"success" example:"true"`
	Data    T      `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// PaginationMeta เป็นโครงสร้างข้อมูลสำหรับส่งกลับข้อมูล Pagination
type PaginationMeta struct {
	TotalItems  int `json:"total_items" example:"100"`
	TotalPages  int `json:"total_pages" example:"10"`
	CurrentPage int `json:"current_page" example:"1"`
	PageSize    int `json:"page_size" example:"10"`
}

// PaginatedResponse เป็นโครงสร้างข้อมูลสำหรับส่งกลับข้อมูล Pagination
type PaginatedResponse[T any] struct {
	Success bool           `json:"success" example:"true"`
	Data    []T            `json:"data"`
	Meta    PaginationMeta `json:"meta"`
	Message string         `json:"message,omitempty"`
}

// --- Helper Functions เพื่อให้ Handler เรียกใช้งานได้ง่ายขึ้น ---

// SuccessResponse สร้าง JsonResponse สำหรับการทำงานสำเร็จ
func Success[T any](c *fiber.Ctx, statusCde int, message string, data T) error {
	return c.Status(statusCde).JSON(JsonResponse[T]{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// OkMessage สร้าง JsonResponse สำหรับการทำงานสำเร็จ โดยไม่มีข้อมูล Data
func OkMessage(c *fiber.Ctx, statusCde int, message string) error {
	return c.Status(statusCde).JSON(JsonResponse[any]{
		Success: true,
		Message: message,
		Data:    nil,
	})
}

// Paginated ส่งกลับสำหรับข้อมูลแบบ List ที่มี Pagination
func Paginated[T any](c *fiber.Ctx, statusCode int, message string, data []T, meta PaginationMeta) error {
	return c.Status(statusCode).JSON(PaginatedResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}
