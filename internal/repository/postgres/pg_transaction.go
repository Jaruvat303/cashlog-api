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
	err := g.db.WithContext(ctx).First(&tx, id).Error
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

	type QueryResult struct {
		CategoryID      *int64
		CategoryName    string
		IconURL         *string
		TransactionType *string
		TotalAmount     float64
	}

	var result []QueryResult

	// ถ้าไม่มี Icon Url ให้ใช้ค่าเริ่มต้น
	defaultIcon := "folder"
	// Query เดียวที่ดึงข้อมูลสรุปตามหมวดหมู่ในช่วงเวลาที่กำหนด
	err := g.db.WithContext(ctx).
		Table("transactions").
		Select(`
		transactions.category_id,
		COALESCE(categories.name, 'Uncategorized') as category_name, 
        COALESCE(categories.icon_url, ?) as icon_url,
		transactions.transaction_type,
		SUM(transactions.amount)as total_amount
		`, defaultIcon).
		Joins("LEFT JOIN categories ON transactions.category_id = categories.id").
		Where("transactions.transaction_date BETWEEN ? AND ?", startDate, endDate).
		Group("transactions.category_id,categories.name,categories.icon_url,transactions.transaction_type").
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
			IconURl:      res.IconURL,
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
func (g *gormTransactionRepository) FetchByTimeRange(ctx context.Context, startDate time.Time, endDate time.Time) ([]domain.Transaction, error) {
	var txs []domain.Transaction

	// SQL: SELECT * FROM transactions WHERE transaction_at BETWEEN ? AND ? ORDER BY transaction_at DESC
	err := g.db.WithContext(ctx).
		Table("transactions").
		Where("transaction_date BETWEEN ? AND ?", startDate, endDate).
		Order("transaction_id DESC").
		Find(&txs).Error

	if err != nil {
		return nil, HandlerDBError(ctx, err, g.log)
	}
	return txs, nil
}

// CheckDuplicate implements [domain.TransactionRepository]. ทำหน้าที่ตรวจสอบรหัสสลิป TransactionID ว่าบันทึกข้อมูลไปแล้วรึยัง
func (g *gormTransactionRepository) CheckDuplicate(ctx context.Context, txID string) (bool, error) {

	var count int64
	// ติวรี่เพื่อนับจำนวนแถวที่มี transaction_id ตรงกับสลืปใบนี้
	err := g.db.WithContext(ctx).
		Model(&domain.Transaction{}).
		Where("transaction_id = ?", txID).
		Count(&count).
		Error

	if err != nil {
		return false, HandlerDBError(ctx, err, g.log)
	}

	// ถ้า count > 0 แสดงว่ามีข้อมูลแล้ว (ซ้ำ)
	return count > 0, nil
}

// Insert implements [domain.TransactionRepository]. บันทึกรายการลงในฐานข้อมูล
func (g *gormTransactionRepository) Insert(ctx context.Context, tx *domain.Transaction) error {
	// ใช้ GORM บันทึกข้อมูลโตรงสร้าง Entity ลงตาราง transaction อัตโนมัติ
	err := g.db.WithContext(ctx).Create(tx).Error
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
