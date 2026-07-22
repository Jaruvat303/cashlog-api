package dto

import "github.com/Jaruvat303/cashlog/internal/domain"

type DashboardSummaryResponse struct {
	TotalIncome  float64                `json:"total_income"`
	TotalExpense float64                `json:"total_expense"`
	Scope        string                 `json:"scope"`
	Month        int                    `json:"month"`
	Year         int                    `json:"year"`
	Income       []CategoryBreakdownDTO `json:"income"`
	Expense      []CategoryBreakdownDTO `json:"expense"`
}

type CategoryBreakdownDTO struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	IconKey      string  `json:"icon_key"`  // เก็บชื่อกลาง เช่น "utensils", "zap", "wallet"
	ColorHex     string  `json:"color_hex"` // เก็บโค้ดสี เช่น "#EF4444"
	TotalAmount  float64 `json:"total_amount"`
}

// MapToDashboardSummaryResponse แปลงข้อมูลจาก Domain Model เป็น DTO สำหรับส่งกลับไปให้ Client
func MapToDashboardSummaryResponse(summary *domain.DashboardSummary) DashboardSummaryResponse {
	if summary == nil {
		return DashboardSummaryResponse{}
	}

	return DashboardSummaryResponse{
		TotalIncome:  summary.TotalIncome,
		TotalExpense: summary.TotalExpense,
		Scope:        summary.Scope,
		Month:        summary.Month,
		Year:         summary.Year,
		Income:       MapToCategoryBreakdownDTO(summary.Income),
		Expense:      MapToCategoryBreakdownDTO(summary.Expense),
	}
}

// MapToCategoryBreakdownDTO แปลงข้อมูลจาก Domain Model เป็น DTO สำหรับส่งกลับไปให้ Client
func MapToCategoryBreakdownDTO(breakdown []domain.CategoryBreakdown) []CategoryBreakdownDTO {
	result := make([]CategoryBreakdownDTO, len(breakdown))
	for i, bd := range breakdown {
		result[i] = CategoryBreakdownDTO{
			CategoryID:   bd.CategoryID,
			CategoryName: bd.CategoryName,
			IconKey:      bd.IconKey,
			ColorHex:     bd.ColorHex,
			TotalAmount:  bd.TotalAmount,
		}
	}
	return result
}
