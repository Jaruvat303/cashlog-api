package dto

// ErrorResponseDTO เป็นโครงสร้างข้อมูลสำหรับส่งกลับเมื่อเกิดข้อผิดพลาด
type ErrorResponseDTO struct {
	Success   bool   `json:"success" example:"false"`
	ErrorCode string `json:"error_code" example:"INVALID_REQUEST_BODY"`
	Message   string `json:"message" example:"ข้อมูลที่ส่งมาไม่ถูกต้องตามรูปแบบ"`
}
