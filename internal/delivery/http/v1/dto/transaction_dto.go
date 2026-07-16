package dto

import (
	"time"

	"github.com/Jaruvat303/cashlog/internal/domain"
)

// TransactionCategoryInfo โครงสร้างข้อมูลหมวดหมู่ย่อส่วนสำหรับแนบไปกับธุรกรรม
type TransactionCategoryInfo struct {
	ID       int64  `json:"id" example:"3"`
	Name     string `json:"name" example:"อาหารและเครื่องดื่ม"`
	Type     string `json:"type" example:"expense"`
	IconKey  string `json:"icon_key" example:"utensils"`
	ColorHex string `json:"color_hex" example:"#EF4444"`
}

// TransactionResponse โครงสร้างข้อมูลหลักสำหรับส่งกลับไปให้ Client / หน้าบ้าน
type TransactionResponse struct {
	ID              uint                     `json:"id" example:"102"`
	Amount          float64                  `json:"amount" example:"350.50"`
	TransactionType string                   `json:"transaction_type" example:"expense"`
	ReceiverName    string                   `json:"receiver_name" example:"ร้านข้าวมันไก่ป้าใจ"`
	Note            string                   `json:"note" example:"มื้อเที่ยงกับทีมงาน"`
	LocalImageName  string                   `json:"local_image_name" example:"slip_20260714_xyz.jpg"`
	TransactionDate string                   `json:"transaction_date" example:"2026-07-14T12:30:00+07:00"`
	Category        *TransactionCategoryInfo `json:"category"` // แนบข้อมูลหมวดหมู่ย่อยแบบ Nested Object
}

// UpdateTransactionInput คือ DTO สำหรับล็อกขอบเขตการแก้ไขข้อมูลจากหน้าบ้าน
type UpdateTransactionInput struct {
	Amount          *float64   `json:"amount" validate:"omitempty,gt=0"`
	Note            *string    `json:"note" validate:"omitempty,max=255"`
	CategoryID      *int64     `json:"category_id" validate:"omitempty,gt=0"`
	TransactionDate *time.Time `json:"transaction_date" validate:"omitempty"`
}

// --- Mapper Functions (แปลงจาก Domain Model เข้าสู่ DTO ขาออก) ---

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

// MapToTransactionResponse แปลงข้อมูลจาก Domain Model เป็น DTO สำหรับส่งกลับไปให้ Client
func MapToTransactionResponse(tx *domain.Transaction) TransactionResponse {

	var categoryInfo *TransactionCategoryInfo

	if tx.Category.ID != 0 {
		categoryInfo = &TransactionCategoryInfo{
			ID:       tx.Category.ID,
			Name:     tx.Category.Name,
			Type:     tx.Category.Type,
			IconKey:  tx.Category.IconKey,
			ColorHex: tx.Category.ColorHex,
		}
	}

	return TransactionResponse{
		ID:              tx.ID,
		Amount:          tx.Amount,
		TransactionType: tx.TransactionType,
		ReceiverName:    tx.ReceiverName,
		Note:            tx.Note,
		LocalImageName:  tx.LocalImageName,
		TransactionDate: tx.TransactionDate.Format(time.RFC3339),
		Category:        categoryInfo,
	}
}

// MapToTransactionListResponse แปลงข้อมูลจาก Domain Model เป็น DTO สำหรับส่งกลับไปให้ Client (แบบ List)
func MapToTransactionListResponse(transactions []domain.Transaction) []TransactionResponse {
	responses := make([]TransactionResponse, len(transactions))

	for i := range transactions {
		responses[i] = MapToTransactionResponse(&transactions[i])
	}

	return responses
}
