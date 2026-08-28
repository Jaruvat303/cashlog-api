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
	SenderName      string                   `json:"sender_name" example:"นายเอ สมมติ"`
	ReceiverName    string                   `json:"receiver_name" example:"ร้านข้าวมันไก่ป้าใจ"`
	Note            string                   `json:"note" example:"มื้อเที่ยงกับทีมงาน"`
	AccountID       *int64                   `json:"account_id" example:"1"`
	FromAccountID   *int64                   `json:"from_account_id" example:"1"`
	ToAccountID     *int64                   `json:"to_account_id" example:"2"`
	Source          string                   `json:"source" example:"slip"`
	LocalImageName  string                   `json:"local_image_name" example:"slip_20260714_xyz.jpg"`
	TransactionDate string                   `json:"transaction_date" example:"2026-07-14T12:30:00+07:00"`
	Category        *TransactionCategoryInfo `json:"category"` // แนบข้อมูลหมวดหมู่ย่อยแบบ Nested Object
}

// UpdateTransactionInput คือ DTO สำหรับล็อกขอบเขตการแก้ไขข้อมูลจากหน้าบ้าน
//
// AccountID ใช้ได้เฉพาะธุรกรรม income/expense — สำหรับ "เติม" account_id ที่ BR-2 auto-scan match ไม่เจอ (nil)
// FromAccountID/ToAccountID ใช้ได้เฉพาะธุรกรรม transfer — สำหรับ "เติม" to_account_id ที่ BR-2 ไม่พยายาม match ตาม Decision #23
type UpdateTransactionInput struct {
	Amount          *float64   `json:"amount" validate:"omitempty,gt=0" example:"150.50"`
	Note            *string    `json:"note" validate:"omitempty,max=255" example:"ค่ากาแฟอเมริกาโน่เย็น"`
	CategoryID      *int64     `json:"category_id" validate:"omitempty,gt=0" example:"2"`
	AccountID       *int64     `json:"account_id" validate:"omitempty,gt=0" example:"1"`
	FromAccountID   *int64     `json:"from_account_id" validate:"omitempty,gt=0" example:"1"`
	ToAccountID     *int64     `json:"to_account_id" validate:"omitempty,gt=0" example:"2"`
	TransactionDate *time.Time `json:"transaction_date" validate:"omitempty" example:"2026-07-24T14:30:00+07:00"`
}

// CreateTransactionInput คือ DTO สำหรับสร้างธุรกรรม income/expense แบบ manual (ไม่ผ่านสลิป)
type CreateTransactionInput struct {
	Amount          float64    `json:"amount" validate:"required,gt=0" example:"150.50"`
	TransactionType string     `json:"transaction_type" validate:"required,oneof=income expense" example:"expense"`
	AccountID       int64      `json:"account_id" validate:"required,gt=0" example:"1"`
	CategoryID      *int64     `json:"category_id" validate:"omitempty,gt=0" example:"2"`
	Note            string     `json:"note" validate:"omitempty,max=255" example:"ค่ากาแฟอเมริกาโน่เย็น"`
	TransactionDate *time.Time `json:"transaction_date" validate:"omitempty" example:"2026-07-24T14:30:00+07:00"`
}

// CreateTransferInput คือ DTO สำหรับสร้างธุรกรรม transfer แบบ manual ระหว่าง 2 บัญชี
// CategoryID มีไว้เพื่อรับค่าจาก client ที่อาจส่งมาแบบผิดสัญญา แล้วถูก reject ที่ usecase layer (Decision #21) — ไม่มีไว้ให้ใช้จริง
type CreateTransferInput struct {
	Amount          float64    `json:"amount" validate:"required,gt=0" example:"150.50"`
	FromAccountID   int64      `json:"from_account_id" validate:"required,gt=0" example:"1"`
	ToAccountID     int64      `json:"to_account_id" validate:"required,gt=0" example:"2"`
	CategoryID      *int64     `json:"category_id" validate:"omitempty,gt=0" example:"2"`
	Note            string     `json:"note" validate:"omitempty,max=255" example:"โอนเงินไปบัญชีออมทรัพย์"`
	TransactionDate *time.Time `json:"transaction_date" validate:"omitempty" example:"2026-07-24T14:30:00+07:00"`
}

// ToDomainCreateParam แปลงข้อมูลจาก DTO เป็น Domain Param
func (c *CreateTransactionInput) ToDomainCreateParam() domain.CreateTransactionParam {
	return domain.CreateTransactionParam{
		Amount:          c.Amount,
		TransactionType: c.TransactionType,
		AccountID:       c.AccountID,
		CategoryID:      c.CategoryID,
		Note:            c.Note,
		TransactionDate: c.TransactionDate,
	}
}

// ToDomainCreateParam แปลงข้อมูลจาก DTO เป็น Domain Param
func (c *CreateTransferInput) ToDomainCreateParam() domain.CreateTransferParam {
	return domain.CreateTransferParam{
		Amount:          c.Amount,
		FromAccountID:   c.FromAccountID,
		ToAccountID:     c.ToAccountID,
		CategoryID:      c.CategoryID,
		Note:            c.Note,
		TransactionDate: c.TransactionDate,
	}
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

	if u.AccountID != nil {
		param.AccountID = u.AccountID
	}

	if u.FromAccountID != nil {
		param.FromAccountID = u.FromAccountID
	}

	if u.ToAccountID != nil {
		param.ToAccountID = u.ToAccountID
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
		SenderName:      tx.SenderName,
		ReceiverName:    tx.ReceiverName,
		Note:            tx.Note,
		AccountID:       tx.AccountID,
		FromAccountID:   tx.FromAccountID,
		ToAccountID:     tx.ToAccountID,
		Source:          tx.Source,
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
