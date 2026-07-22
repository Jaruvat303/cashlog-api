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
)
