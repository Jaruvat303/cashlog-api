package domain

// DashboardSummary โตรงสร้างข้อมูลสรุปผลรายรับรายจ่าน
type DashboardSummary struct {
	TotalIncome   float64             `json:"total_income"`
	TotalExpense  float64             `json:"total_expense"`
	TotalTransfer float64             `json:"total_transfer"`
	Scope         string              `json:"scope"`
	Month         int                 `json:"month"`
	Year          int                 `json:"year"`
	Income        []CategoryBreakdown `json:"income"`
	Expense       []CategoryBreakdown `json:"expense"`
}

// โครงสร้างย่อย: ยอดรวมแยกตามแต่ละหมวดหมู่
type CategoryBreakdown struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	IconKey      string  `json:"icon_key"`  // เก็บชื่อกลาง เช่น "utensils", "zap", "wallet"
	ColorHex     string  `json:"color_hex"` // เก็บโค้ดสี เช่น "#EF4444"
	TotalAmount  float64 `json:"total_amount"`
}
