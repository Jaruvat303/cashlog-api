package domain

// ค่าที่เป็นไปได้ของ granularity สำหรับ Trend endpoint
const (
	TrendGranularityMonth = "month"
	TrendGranularityYear  = "year"
)

// TrendBucket คือช่องข้อมูลหนึ่งช่องของกราฟแท่งรายรับ-รายจ่าย (หนึ่งเดือนหรือหนึ่งปี)
type TrendBucket struct {
	Year         int
	Month        *int // nil เมื่อ granularity เป็น year
	TotalIncome  float64
	TotalExpense float64
	Net          float64
}

// TrendResult คือผลลัพธ์ทั้งชุดที่ UseCase ส่งกลับให้ Handler
type TrendResult struct {
	Granularity string
	Year        *int // ไม่เป็น nil เฉพาะตอน granularity เป็น month
	Buckets     []TrendBucket
}

// MonthlyAggregate คือผลรวมรายรับ-รายจ่ายของหนึ่งเดือน (ตามเวลา Asia/Bangkok) ที่ได้จาก Repository
// UseCase เป็นคนตัดสินใจเองว่าจะใช้ตรงๆ (month mode) หรือ roll up รวมเป็นรายปี (year mode) —
// granularity ไม่ไหลลงไปถึงชั้น Repository/SQL เลย (Ticket B2)
type MonthlyAggregate struct {
	Year         int
	Month        int // 1-12
	TotalIncome  float64
	TotalExpense float64
}
