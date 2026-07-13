package dto

import (
	"time"

	"github.com/Jaruvat303/cashlog/internal/domain"
)

// UpdateTransactionInput คือ DTO สำหรับล็อกขอบเขตการแก้ไขข้อมูลจากหน้าบ้าน
type UpdateTransactionInput struct {
	Amount          *float64   `json:"amount" validate:"omitempty,gt=0"`
	Note            *string    `json:"note" validate:"omitempty,max=255"`
	CategoryID      *int64     `json:"category_id" validate:"omitempty,gt=0"`
	TransactionDate *time.Time `json:"transaction_date" validate:"omitempty"`
}

// ToDomainUpdateParam แปลงข้อมูลจาก DTO เป็น Domain Param
func (u *UpdateTransactionInput) ToDomainUpdateParam() domain.UpdateTransactionParam {
	param := domain.UpdateTransactionParam{}

	if u.Amount != nil {
		param.Amount = u.Amount
	}

	if u.Note != nil {
		param.Note = u.Note
	}

	if u.CategoryID != nil {
		param.CategoryID = u.CategoryID
	}

	if u.TransactionDate != nil {
		param.TransactionDate = u.TransactionDate
	}

	return param
}
