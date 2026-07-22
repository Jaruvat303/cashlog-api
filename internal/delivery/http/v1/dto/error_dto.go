package dto

// ErrorResponseDTO เป็นโครงสร้างข้อมูลสำหรับส่งกลับเมื่อเกิดข้อผิดพลาด
type ErrorResponseDTO struct {
	Success   bool   `json:"success" example:"false"`
	ErrorCode string `json:"error_code" example:"ERROR_CODE_HERE"`
	Message   string `json:"message" example:"ข้อความอธิบายความผิดพลาดตามสถานการณ์"`
}

// EmptyData โครงสร้างว่างสำหรับบอก Swagger ว่าไม่มีข้อมูลส่งกลับ (จะแสดงผลเป็น data: {})
type EmptyData struct{}
