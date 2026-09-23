package dto

import "github.com/Jaruvat303/cashlog/internal/domain"

type TrendResponse struct {
	Granularity string           `json:"granularity"`
	Year        *int             `json:"year"`
	Buckets     []TrendBucketDTO `json:"buckets"`
}

type TrendBucketDTO struct {
	Year         int     `json:"year"`
	Month        *int    `json:"month"`
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Net          float64 `json:"net"`
}

// MapToTrendResponse แปลงข้อมูลจาก Domain Model เป็น DTO สำหรับส่งกลับไปให้ Client
func MapToTrendResponse(result *domain.TrendResult) TrendResponse {
	if result == nil {
		return TrendResponse{}
	}

	buckets := make([]TrendBucketDTO, len(result.Buckets))
	for i, b := range result.Buckets {
		buckets[i] = TrendBucketDTO{
			Year:         b.Year,
			Month:        b.Month,
			TotalIncome:  b.TotalIncome,
			TotalExpense: b.TotalExpense,
			Net:          b.Net,
		}
	}

	return TrendResponse{
		Granularity: result.Granularity,
		Year:        result.Year,
		Buckets:     buckets,
	}
}
