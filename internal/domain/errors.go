package domain

import "errors"

// ประกาศตัวแปร Error ที่อาจเกิดขึ้นได้ในระบบธุรกิจของเรา
var (
	ErrNotFound             = errors.New("data not found")
	ErrDuplicateRequest     = errors.New("duplicate request detected")
	ErrInvalidInput         = errors.New("invalid input parameters")
	ErrInternalDB           = errors.New("internal database error")
	ErrContextCanceled      = errors.New("request canceled by user")
	ErrTimeout              = errors.New("database query timeout exceeded")
	ErrGeminiQuotaExhausted = errors.New("gemini_quota_exhausted") // Billing หมด / ติด Quota 429
	ErrGeminiUnavailable    = errors.New("gemini_unavailable")     // Gemini Service ล่ม / ติด 5xx
	ErrGeminiEmptyResponse  = errors.New("gemini_empty_response")  // AI ไม่ตอบกลับมา
	ErrSlipParseFailed      = errors.New("slip_parse_failed")      // Unmarshal JSON จาก AI ไม่ผ่าน

	ErrAccountInactive               = errors.New("account_inactive")                  // อ้างอิง account ที่ is_active=false
	ErrTransferSameAccount           = errors.New("transfer_same_account")             // from_account_id == to_account_id
	ErrCategoryNotAllowedForTransfer = errors.New("category_not_allowed_for_transfer") // ส่ง category_id มาตอนสร้าง transfer
)
