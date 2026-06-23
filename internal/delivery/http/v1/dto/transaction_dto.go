package dto

import "time"

// UpdateTransactionInput คือ DTO สำหรับล็อกขอบเขตการแก้ไขข้อมูลจากหน้าบ้าน
type UpdateTransactionInput struct {
	Amount          *float64   `json:"amount" validate:"omitempty,gt=0"`
	Note            *string    `json:"note" validate:"omitempty,max=255"`
	CategoryID      *int64     `json:"category_id" validate:"omitempty,gt=0"`
	TransactionDate *time.Time `json:"transaction_date" validate:"omitempty"`
}
