package domain

import (
	"context"
	"time"
)

type Transaction struct {
	ID              uint
	Amount          float64
	TransactionType string
	ReceiverName    string
	Note            string
	CategoryID      int64
	LocalImageName  string
	TransactionDate time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TransactionRepository interface {
	Insert(ctx context.Context, tx *Transaction) error
	FetchByTimeRange(ctx context.Context, startDate, endDate time.Time) ([]Transaction, error)
	CalculateSummary(ctx context.Context, startDate, endDate time.Time, scope string) (*DashboardSummary, error)
	Update(ctx context.Context, tx *Transaction) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*Transaction, error)
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
	GetMonthlyHistory(ctx context.Context, month, year int) ([]Transaction, error)
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
