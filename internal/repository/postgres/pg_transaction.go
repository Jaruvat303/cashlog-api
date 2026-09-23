package postgres

import (
	"context"
	"time"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormTransactionRepository struct {
	db  *gorm.DB
	log logger.Logger
}

// CountByTimeRange implements [domain.TransactionRepository].
func (g *gormTransactionRepository) CountByTimeRange(ctx context.Context, startDate time.Time, endDate time.Time) (int64, error) {
	var count int64
	err := g.db.WithContext(ctx).
		Model(&domain.Transaction{}).
		Where("transaction_date BETWEEN ? AND ?", startDate, endDate).
		Count(&count).Error

	if err != nil {
		return 0, HandlerDBError(ctx, err, g.log)
	}
	return count, nil
}

// Delete implements [domain.TransactionRepository].
func (g *gormTransactionRepository) Delete(ctx context.Context, id uint) error {
	err := g.db.WithContext(ctx).Delete(&domain.Transaction{}, id).Error
	if err != nil {
		return HandlerDBError(ctx, err, g.log)
	}
	return nil
}

// GetByID implements [domain.TransactionRepository].
func (g *gormTransactionRepository) GetByID(ctx context.Context, id uint) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := g.db.WithContext(ctx).Preload("Category").First(&tx, id).Error
	if err != nil {
		return nil, HandlerDBError(ctx, err, g.log)
	}
	return &tx, nil
}

// update implements [domain.TransactionRepository].
func (g *gormTransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	// Omit(clause.Associations): tx.Category/Account/FromAccount/ToAccount อาจยังมีค่าเก่าค้างอยู่จากตอน GetByID Preload
	// ถ้าไม่ Omit, GORM Save จะ sync FK (เช่น category_id) กลับจาก association struct ที่ preload ไว้ ทับค่าที่ Usecase เพิ่งเคลียร์เป็น nil (Ticket 04 type conversion)
	err := g.db.WithContext(ctx).Omit(clause.Associations).Save(tx).Error
	if err != nil {
		return HandlerDBError(ctx, err, g.log)
	}

	// รีเซ็ต struct ทั้งก้อนก่อน reload ใหม่: ถ้าไม่รีเซ็ต ตอน category_id เป็น nil (เช่นหลังแปลงเป็น transfer)
	// GORM Preload จะไม่มีอะไรให้ preload เลยปล่อย tx.Category ที่ preload ค้างมาจาก GetByID ไว้เหมือนเดิม ทำให้ response หลอกว่ายังมี category ติดอยู่
	id := tx.ID
	*tx = domain.Transaction{}
	// สั่งโหลดข้อมูลใหม่ล่าสุดจาก DB พ่วง Category กลับมาส่งให้ชั้นนอก
	err = g.db.WithContext(ctx).Preload("Category").First(tx, id).Error
	if err != nil {
		return HandlerDBError(ctx, err, g.log)
	}
	return nil
}

// CalculateSummary implements [domain.TransactionRepository].
func (g *gormTransactionRepository) CalculateSummary(ctx context.Context, startDate time.Time, endDate time.Time, scope string) (*domain.DashboardSummary, error) {
	summary := &domain.DashboardSummary{
		Scope:   scope,
		Month:   int(startDate.Month()),
		Year:    startDate.Year(),
		Income:  []domain.CategoryBreakdown{},
		Expense: []domain.CategoryBreakdown{},
	}

	// QueryResult เป็นโครงสร้างชั่วคราวสำหรับเก็บผลลัพธ์จาก Query
	type QueryResult struct {
		CategoryID      *int64
		CategoryName    string
		IconKey         *string
		ColorHex        *string
		TransactionType *string
		TotalAmount     float64
	}

	var result []QueryResult

	// ถ้าไม่มี Icon Url ให้ใช้ค่าเริ่มต้น
	defaultIcon := "folder"
	defaultColor := "#CCCCCC"
	// Query เดียวที่ดึงข้อมูลสรุปตามหมวดหมู่ในช่วงเวลาที่กำหนด
	err := g.db.WithContext(ctx).
		Table("transactions").
		Select(`
		transactions.category_id,
		COALESCE(categories.name, 'Uncategorized') as category_name,
        COALESCE(categories.icon_key, ?) as icon_key,
		COALESCE(categories.color_hex, ?) as color_hex,
		transactions.transaction_type,
		SUM(transactions.amount)as total_amount
		`, defaultIcon, defaultColor).
		Joins("LEFT JOIN categories ON transactions.category_id = categories.id").
		Where("transactions.transaction_date BETWEEN ? AND ?", startDate, endDate).
		Group("transactions.category_id,categories.name,categories.icon_key,categories.color_hex,transactions.transaction_type").
		Scan(&result).Error

	if err != nil {
		return nil, HandlerDBError(ctx, err, g.log)
	}

	for _, res := range result {
		// กรณี CategoryID เป็น Null ให้เป็นค่าเริ่มต้น
		var catID int64
		if res.CategoryID != nil {
			catID = *res.CategoryID
		}

		breakdown := domain.CategoryBreakdown{
			CategoryID:   catID,
			CategoryName: res.CategoryName,
			IconKey:      defaultIcon,
			ColorHex:     defaultColor,
			TotalAmount:  res.TotalAmount,
		}

		if res.IconKey != nil {
			breakdown.IconKey = *res.IconKey
		}
		if res.ColorHex != nil {
			breakdown.ColorHex = *res.ColorHex
		}

		txType := ""
		if res.TransactionType != nil {
			txType = *res.TransactionType
		}

		// แยกประเภทรายรับ/รายจ่าย/เงินโอนย้ายระหว่างบัญชี (BR-4)
		// เดิมเทียบ "INCOME" (uppercase) แต่ค่าจริงเก็บเป็น lowercase ("income") ทำให้ทุกรายการตกไปเป็น expense หมด (bug #12)
		switch txType {
		case domain.TransactionTypeIncome:
			summary.TotalIncome += res.TotalAmount
			summary.Income = append(summary.Income, breakdown)
		case domain.TransactionTypeTransfer:
			// transfer ไม่มี category (category_id เป็น nil เสมอ) จึงไม่เก็บ breakdown แยกหมวดหมู่
			summary.TotalTransfer += res.TotalAmount
		default:
			summary.TotalExpense += res.TotalAmount
			summary.Expense = append(summary.Expense, breakdown)
		}
	}

	return summary, nil
}

// AggregateMonthly implements [domain.TransactionRepository]. รวมยอดรายรับ-รายจ่ายแบบ group รายเดือน
// ตามเวลา Asia/Bangkok ภายในขอบเขต [from, to) — filter บน transaction_date เป็น sargable range
// (>= AND <) ตรงๆ เพื่อให้ idx_transactions_transaction_date ทำงาน (Ticket B2)
// ใช้ .Model(&domain.Transaction{}) แทนการ hardcode ชื่อตาราง เพื่อให้ table name resolve ถูกเสมอ
// (dev/prod แยกกันด้วย search_path บน DSN ไม่ใช่ GORM TablePrefix แต่กัน regression ไว้ก่อน)
func (g *gormTransactionRepository) AggregateMonthly(ctx context.Context, from, to time.Time) ([]domain.MonthlyAggregate, error) {
	var result []domain.MonthlyAggregate

	err := g.db.WithContext(ctx).
		Model(&domain.Transaction{}).
		Select(`
		EXTRACT(YEAR FROM date_trunc('month', transaction_date AT TIME ZONE 'Asia/Bangkok'))::int AS year,
		EXTRACT(MONTH FROM date_trunc('month', transaction_date AT TIME ZONE 'Asia/Bangkok'))::int AS month,
		SUM(CASE WHEN transaction_type = ? THEN amount ELSE 0 END) AS total_income,
		SUM(CASE WHEN transaction_type = ? THEN amount ELSE 0 END) AS total_expense
		`, domain.TransactionTypeIncome, domain.TransactionTypeExpense).
		Where("transaction_date >= ? AND transaction_date < ?", from, to).
		Group("1, 2").
		Order("1, 2").
		Scan(&result).Error

	if err != nil {
		return nil, HandlerDBError(ctx, err, g.log)
	}
	return result, nil
}

// GetFirstTransactionYear implements [domain.TransactionRepository]. คืนปี (Asia/Bangkok) ของธุรกรรม
// ที่เก่าที่สุดในระบบ โดยห่อ MIN(transaction_date) ด้วย AT TIME ZONE ก่อนตัดปี เพื่อให้ Postgres ยังใช้
// idx_transactions_transaction_date ทำ index scan ของ MIN ได้ (Ticket B2)
func (g *gormTransactionRepository) GetFirstTransactionYear(ctx context.Context) (int, bool, error) {
	var year *int

	err := g.db.WithContext(ctx).
		Model(&domain.Transaction{}).
		Select("EXTRACT(YEAR FROM (MIN(transaction_date) AT TIME ZONE 'Asia/Bangkok'))::int").
		Scan(&year).Error

	if err != nil {
		return 0, false, HandlerDBError(ctx, err, g.log)
	}
	if year == nil {
		return 0, false, nil
	}
	return *year, true, nil
}

// FetchByTimeRange implements [domain.TransactionRepository].
func (g *gormTransactionRepository) FetchByTimeRange(ctx context.Context, param domain.QueryTransactionParam) ([]domain.Transaction, error) {
	var txs []domain.Transaction

	err := g.db.WithContext(ctx).
		Table("transactions").
		Preload("Category").
		Where("transaction_date BETWEEN ? AND ?", param.StartDate, param.EndDate).
		Order("transaction_date DESC,id DESC").Find(&txs).Error

	if err != nil {
		return nil, HandlerDBError(ctx, err, g.log)
	}
	return txs, nil
}

// Insert implements [domain.TransactionRepository]. บันทึกรายการลงในฐานข้อมูล
func (g *gormTransactionRepository) Insert(ctx context.Context, tx *domain.Transaction) error {
	// ใช้ GORM บันทึกข้อมูลโตรงสร้าง Entity ลงตาราง transaction อัตโนมัติ
	err := g.db.WithContext(ctx).Create(tx).Error
	if err != nil {
		return HandlerDBError(ctx, err, g.log)
	}

	// เติมเต็มข้อมูลความสัมพันธ์ (Refresh Data) ด้วยการ Preload Category กลับมาใส่ใน Pointer ตัวเดิม
	err = g.db.WithContext(ctx).Preload("Category").First(tx, tx.ID).Error
	if err != nil {
		return HandlerDBError(ctx, err, g.log)
	}
	return nil
}

// NewGormTransactionRepository สำหรับสร้างอินสแตนซ์สำหรับจัดการฐานข้อมูล Postgres
func NewGormTransactionRepository(db *gorm.DB, appLogger logger.Logger) domain.TransactionRepository {
	return &gormTransactionRepository{
		db:  db,
		log: appLogger,
	}
}
