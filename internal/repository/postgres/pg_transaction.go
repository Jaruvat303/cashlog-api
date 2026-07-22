package postgres

import (
	"context"
	"time"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"gorm.io/gorm"
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
	err := g.db.WithContext(ctx).Save(tx).Error
	if err != nil {
		return HandlerDBError(ctx, err, g.log)
	}

	// สั่งโหลดข้อมูลใหม่ล่าสุดจาก DB พ่วง Category กลับมาส่งให้ชั้นนอก
	err = g.db.WithContext(ctx).Preload("Category").First(tx, tx.ID).Error
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

		summary.TotalIncome += 0
		summary.TotalExpense += 0

		breakdown = domain.CategoryBreakdown{
			CategoryID:   catID,
			CategoryName: res.CategoryName,
			IconKey:      breakdown.IconKey,
			ColorHex:     breakdown.ColorHex,
			TotalAmount:  res.TotalAmount,
		}

		// แยกประเภทรายรับ รายจ่าย
		if res.TransactionType != nil && *res.TransactionType == "INCOME" {
			summary.TotalIncome += res.TotalAmount
			summary.Income = append(summary.Income, breakdown)
		} else {
			summary.TotalExpense += res.TotalAmount
			summary.Expense = append(summary.Expense, breakdown)
		}
	}

	return summary, nil
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
