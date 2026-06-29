package domain

// DashboardSummary โตรงสร้างข้อมูลสรุปผลรายรับรายจ่าน
type DashboardSummary struct {
	TotalIncome  float64             `json:"total_income"`
	TotalExpense float64             `json:"total_expense"`
	Scope        string              `json:"scope"`
	Month        int                 `json:"month"`
	Year         int                 `json:"year"`
	Income       []CategoryBreakdown `json:"income"`
	Expense      []CategoryBreakdown `json:"expense"`
}

// โครงสร้างย่อย: ยอดรวมแยกตามแต่ละหมวดหมู่
type CategoryBreakdown struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	IconURl      *string `json:"icon_url"`
	TotalAmount  float64 `json:"total_amount"`
}
