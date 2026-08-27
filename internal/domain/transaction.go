package domain

import (
	"context"
	"time"
)

// ค่าที่เป็นไปได้ของ TransactionType
const (
	TransactionTypeIncome   = "income"
	TransactionTypeExpense  = "expense"
	TransactionTypeTransfer = "transfer"
)

// ค่าที่เป็นไปได้ของ Source (ที่มาของ transaction)
const (
	TransactionSourceSlip   = "slip"
	TransactionSourceManual = "manual"
)

type Transaction struct {
	ID              uint    `gorm:"primaryKey;autoIncrement"`
	Amount          float64 `gorm:"type:numeric(12,2);not null"`
	TransactionType string  `gorm:"type:varchar(50);not null"` // income, expense, transfer
	SenderName      string  `gorm:"type:varchar(255)"`
	ReceiverName    string  `gorm:"type:varchar(255)"`
	Note            string  `gorm:"type:text"`
	CategoryID      *int64
	Category        Category `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	AccountID       *int64
	Account         *Account `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	FromAccountID   *int64
	FromAccount     *Account `gorm:"foreignKey:FromAccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ToAccountID     *int64
	ToAccount       *Account  `gorm:"foreignKey:ToAccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Source          string    `gorm:"type:varchar(20)"` // slip, manual
	LocalImageName  string    `gorm:"type:varchar(255)"`
	TransactionDate time.Time `gorm:"not null"`
	CreatedAt       time.Time `gorm:"autoCreateTime;not null"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime;not null"`
}

type TransactionRepository interface {
	Insert(ctx context.Context, tx *Transaction) error
	FetchByTimeRange(ctx context.Context, param QueryTransactionParam) ([]Transaction, error)
	CalculateSummary(ctx context.Context, startDate, endDate time.Time, scope string) (*DashboardSummary, error)
	Update(ctx context.Context, tx *Transaction) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Transaction, error)
	CountByTimeRange(ctx context.Context, startDate, endDate time.Time) (int64, error)
}

// CacheRepository
type CacheRepository interface {
	// Data Cache
	GetCache(ctx context.Context, periodKey string) (string, error)
	SetCache(ctx context.Context, periodKey string, jsonData string) error
	InvalidateCache(ctx context.Context, periodKey string) error

	// File/Storage Cache
	CheckFileExists(ctx context.Context, localImageName string) (bool, error)
	SetFileCache(ctx context.Context, localImageName string) error
}

type TransactionUsecase interface {
	// SyncTransaction รับไฟล์ภาพสลิปในรูปแบบ byte array และชื่อไฟล์ภาพ เพื่อไปประมาลผลและบันทึกข้อมูล
	SyncTransaction(ctx context.Context, imageBytes []byte, localImageName string) (*Transaction, error)
	FetchTransactions(ctx context.Context, input FetchTransactionInput) (*FetchTransactionResult, error)
	GetDashboardSummary(ctx context.Context, scope string, month, year int) (*DashboardSummary, error)
	UpdateTransaction(ctx context.Context, id uint, input UpdateTransactionParam) (*Transaction, error)
	DeleteTransaction(ctx context.Context, id uint) error
}

type UpdateTransactionParam struct {
	Amount          *float64
	Note            *string
	CategoryID      *int64
	TransactionDate *time.Time
}

// พารามิเตอร์ที่ UseCase รับมาจาก Handler
type FetchTransactionInput struct {
	Year  int // เช่น 2026 (ถ้าเป็น 0 แปลว่าเอาปีปัจจุบัน)
	Month int // 1-12 (ถ้าเป็น 0 แปลว่าเอาเดือนปัจจุบัน)
	Page  int // หน้าที่ต้องการ เช่น 1
	Limit int // จำนวนต่อหน้า เช่น 20
}

// พารามิเตอร์ที่ส่งไปให้ Repository เพื่อดึงข้อมูล
type QueryTransactionParam struct {
	StartDate time.Time
	EndDate   time.Time
	Limit     int
	Offset    int
}

// สิ่งที่ UseCase ส่งกลับไปให้ Handler (ประกอบด้วยข้อมูล และข้อมูลสรุปจำนวน)
type FetchTransactionResult struct {
	Transactions []Transaction
	TotalItems   int64
}
