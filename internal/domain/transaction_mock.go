package domain

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type TransactionUsecaseMock struct {
	mock.Mock
}

func (m *TransactionUsecaseMock) SyncTransaction(ctx context.Context, imageBytes []byte, localImageName string) (*Transaction, error) {
	args := m.Called(ctx, imageBytes, localImageName)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *TransactionUsecaseMock) CreateTransaction(ctx context.Context, input CreateTransactionParam) (*Transaction, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *TransactionUsecaseMock) CreateTransfer(ctx context.Context, input CreateTransferParam) (*Transaction, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *TransactionUsecaseMock) FetchTransactions(ctx context.Context, input FetchTransactionInput) (*FetchTransactionResult, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FetchTransactionResult), args.Error(1)
}

func (m *TransactionUsecaseMock) GetDashboardSummary(ctx context.Context, scope string, month, year int) (*DashboardSummary, error) {
	args := m.Called(ctx, scope, month, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DashboardSummary), args.Error(1)
}

func (m *TransactionUsecaseMock) UpdateTransaction(ctx context.Context, id uint, input UpdateTransactionParam) (*Transaction, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *TransactionUsecaseMock) DeleteTransaction(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TransactionRepositoryMock จำลอง function ที่มีใน Database

type TransactionRepositoryMock struct {
	mock.Mock
}

func (m *TransactionRepositoryMock) Insert(ctx context.Context, tx *Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *TransactionRepositoryMock) CheckDuplicate(ctx context.Context, txID string) (bool, error) {
	args := m.Called(ctx, txID)
	return args.Bool(0), args.Error(1)
}

func (m *TransactionRepositoryMock) FetchByTimeRange(ctx context.Context, param QueryTransactionParam) ([]Transaction, error) {
	args := m.Called(ctx, param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Transaction), args.Error(1)
}

func (m *TransactionRepositoryMock) CalculateSummary(ctx context.Context, startDate, endDate time.Time, scope string) (*DashboardSummary, error) {
	args := m.Called(ctx, startDate, endDate, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DashboardSummary), args.Error(1)
}

func (m *TransactionRepositoryMock) Update(ctx context.Context, tx *Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *TransactionRepositoryMock) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TransactionRepositoryMock) GetByID(ctx context.Context, id uint) (*Transaction, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *TransactionRepositoryMock) CountByTimeRange(ctx context.Context, startDate, endDate time.Time) (int64, error) {
	args := m.Called(ctx, startDate, endDate)
	return args.Get(0).(int64), args.Error(1)
}

// TransactionCacheRepositoryMock stuct สำหรับจำลอง function ใน Redis Cache

type TransactionCacheRepositoryMock struct {
	mock.Mock
}

func (m *TransactionCacheRepositoryMock) GetCache(ctx context.Context, periodKey string) (string, error) {
	args := m.Called(ctx, periodKey)
	return args.String(0), args.Error(1)
}

func (m *TransactionCacheRepositoryMock) SetCache(ctx context.Context, periodKey string, jsonData string) error {
	args := m.Called(ctx, periodKey, jsonData)
	return args.Error(0)
}

func (m *TransactionCacheRepositoryMock) InvalidateCache(ctx context.Context, periodKey string) error {
	args := m.Called(ctx, periodKey)
	return args.Error(0)
}

func (m *TransactionCacheRepositoryMock) CheckFileExists(ctx context.Context, localImageName string) (bool, error) {
	args := m.Called(ctx, localImageName)
	return args.Bool(0), args.Error(1)
}

func (m *TransactionCacheRepositoryMock) SetFileCache(ctx context.Context, localImageName string) error {
	args := m.Called(ctx, localImageName)
	return args.Error(0)
}
