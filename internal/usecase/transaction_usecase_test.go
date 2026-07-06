package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/internal/usecase"
	"github.com/Jaruvat303/cashlog/pkg"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/Jaruvat303/cashlog/pkg/timeutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMonthlyHistory(t *testing.T) {
	transactionType := "expense"
	// กำหนดวันเริ่มต้นของเดือน
	startDate := time.Date(2026, time.Month(6), 1, 0, 0, 0, 0, timeutil.BangKokLoc)

	// กำหนดวันสุดท้ายของเดือน (บวกไป 1 เดือนแล้วหักออก 1 นาโนวินาที)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	mockMonthlyHistory := []domain.Transaction{
		{
			ID:              1,
			Amount:          2600,
			TransactionType: transactionType,
			LocalImageName:  "image1.jpg",
		},
		{
			ID:              2,
			Amount:          2000,
			TransactionType: transactionType,
			LocalImageName:  "image.jpg",
		},
	}

	tests := []struct {
		name           string
		month          int
		year           int
		setupMock      func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock)
		expectedResult []domain.Transaction
		expectedError  bool
	}{
		{
			name:  "1. Successfully",
			month: 6,
			year:  2026,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				ctx := mock.Anything

				repo.On("FetchByTimeRange", ctx, startDate, endDate).Return(mockMonthlyHistory, nil)
			},
			expectedResult: mockMonthlyHistory,
			expectedError:  false,
		},
		{
			name:  "2. DB Error",
			month: 6,
			year:  2026,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				ctx := mock.Anything
				repo.On("FetchByTimeRange", ctx, startDate, endDate).Return(nil, errors.New("DB Error"))
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(domain.TransactionRepositoryMock)
			mockCacheRepo := new(domain.TransactionCacheRepositoryMock)
			mockGemini := new(domain.GeminiSlipRepositoryMock)
			mockLogger := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo, mockCacheRepo)

			txUsecase := usecase.NewTransactionUsecase(mockRepo, mockCacheRepo, mockGemini, mockLogger)

			// Act
			result, err := txUsecase.GetMonthlyHistory(ctx, tt.month, tt.year)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
			mockCacheRepo.AssertExpectations(t)
		})
	}
}

func TestGetDashboardSummary(t *testing.T) {
	// จำลองข่อมูลที่จะส่งกลับมาจาก Cache หรือ DB
	mockSummary := &domain.DashboardSummary{
		TotalIncome:  15000.0,
		TotalExpense: 5000.0,
		Scope:        "monthly",
		Month:        6,
		Year:         2026,
	}
	mockSummaryJSON, _ := json.Marshal(mockSummary)

	// โตรงสร้าง Table-Driven Test
	tests := []struct {
		name           string
		scope          string
		month          int
		year           int
		setupMock      func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock)
		expectedResult *domain.DashboardSummary
		expectedError  bool
	}{
		{
			name:  "1. Cache Hit - Monthly",
			scope: "monthly",
			month: 6,
			year:  2026,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				ctx := mock.Anything
				periodKey := "summary:monthly:2026-06"

				// จำลองว่า Redis มีข้อมูล (GetCache คืน string JSON,error เป็น nil)
				cache.On("GetCache", ctx, periodKey).Return(string(mockSummaryJSON), nil)

				// พอ Cache Hit แล้ว โค้ดจะไม่วิ่งไปหา Database แน่นอน (ไม่ต้องตั้งค่า repo)
			},
			expectedResult: mockSummary,
			expectedError:  false,
		},
		{
			name:  "2. Cache Miss & Fetch DB Success (Monthly) - ไม่เจอใน Cache แต่ดึงจาก DB ได้ และ Save Cache สำเร็จ",
			scope: "monthly",
			month: 6,
			year:  2026,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				ctx := mock.Anything
				startDate := time.Date(2026, time.Month(6), 1, 0, 0, 0, 0, timeutil.BangKokLoc)
				endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
				periodKey := "summary:monthly:2026-06"

				// จำลองการดึง Cache แล้วไม่เจอ
				cache.On("GetCache", ctx, periodKey).Return("", errors.New("cache miss"))

				// จำลองว่าตำนวนจาก DB สำเร็จ
				repo.On("CalculateSummary", ctx, startDate, endDate, "monthly").Return(mockSummary, nil)

				// บันทึก Cache หลังจากดึง DB เสร็จ
				cache.On("SetCache", ctx, periodKey, mock.Anything).Return(nil)
			},
			expectedResult: mockSummary,
			expectedError:  false,
		},
		{
			name:  "3. Cache Miss & Fatch DB Success - Yearly",
			scope: "yearly",
			month: 1,
			year:  2026,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				ctx := mock.Anything
				periodKey := "summary:year:2026"

				// กำหนดวันเริ่มต้นและวันสุดท้าย
				startDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, timeutil.BangKokLoc)
				endDate := startDate.AddDate(1, 0, 0).Add(-time.Nanosecond)

				yearlySummary := &domain.DashboardSummary{
					TotalIncome:  150000.0,
					TotalExpense: 40000.0,
					Scope:        "yearly",
					Year:         2026,
				}

				cache.On("GetCache", ctx, periodKey).Return("", errors.New("cache miss"))
				repo.On("CalculateSummary", ctx, startDate, endDate, "yearly").Return(yearlySummary, nil)
				cache.On("SetCache", ctx, periodKey, mock.Anything).Return(nil)
			},
			expectedResult: &domain.DashboardSummary{
				TotalIncome:  150000.0,
				TotalExpense: 40000.0,
				Scope:        "yearly",
				Year:         2026,
			},
			expectedError: false,
		},
		{
			name:  "4. Database Failure - ค้นหาจาก DB พัง",
			scope: "monthly",
			month: 6,
			year:  2026,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				ctx := mock.Anything
				periodKey := "summary:monthly:2026-06"

				startDate := time.Date(2026, time.Month(6), 1, 0, 0, 0, 0, timeutil.BangKokLoc)
				endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

				cache.On("GetCache", ctx, periodKey).Return("", errors.New("cache miss"))

				// จำลองการพังของ DB
				repo.On("CalculateSummary", ctx, startDate, endDate, "monthly").Return(nil, errors.New("database connection timeout"))
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	// for loop test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: สร้างวัตถุจำลองและฉีดเข้าไปใน usecase
			mockRepo := new(domain.TransactionRepositoryMock)
			mockCache := new(domain.TransactionCacheRepositoryMock)
			mockGemini := new(domain.GeminiSlipRepositoryMock)
			mockLogger := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo, mockCache)

			logger.InitLogger("development")
			// นำ Mock ไปใส่ใน usecase
			txUsecase := usecase.NewTransactionUsecase(mockRepo, mockCache, mockGemini, mockLogger)

			// Act
			result, err := txUsecase.GetDashboardSummary(ctx, tt.scope, tt.month, tt.year)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Scope, result.Scope)
				assert.Equal(t, tt.expectedResult.TotalIncome, result.TotalIncome)
				assert.Equal(t, tt.expectedResult.TotalExpense, result.TotalExpense)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)

		})
	}
}

func TestDeleteTransaction(t *testing.T) {
	mockFetchTransaction := &domain.Transaction{
		ID:         1,
		Amount:     200,
		Note:       "",
		CategoryID: int64(1),
	}

	tests := []struct {
		name          string
		id            uint
		setupMock     func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock)
		expectedError bool
	}{
		{
			name: "1. DB Error - ค้นหาข้อมูล Transaction ไม่พบ",
			id:   99,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(99)).Return(nil, domain.ErrNotFound)
			},

			expectedError: true,
		},
		{
			name: "2. DB Error - ลบข้อมูลไม่สำเร็จ ฐานข้อมูลมีปัญหา",
			id:   1,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(mockFetchTransaction, nil)
				repo.On("Delete", mock.Anything, uint(1)).Return(domain.ErrInternalDB)
			},
			expectedError: true,
		},
		{
			name: "3. Success - ลบข้อมูลสำเร็จ",
			id:   1,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(mockFetchTransaction, nil)
				repo.On("Delete", mock.Anything, uint(1)).Return(nil)
				cache.On("InvalidateCache", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(domain.TransactionRepositoryMock)
			mockCache := new(domain.TransactionCacheRepositoryMock)
			mockGemini := new(domain.GeminiSlipRepositoryMock)
			mockLogger := logger.NewNopLogger()

			tt.setupMock(mockRepo, mockCache)

			ctx := context.Background()

			uc := usecase.NewTransactionUsecase(mockRepo, mockCache, mockGemini, mockLogger)

			// Act
			err := uc.DeleteTransaction(ctx, tt.id)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)

		})
	}

}

func TestUpdateTransaction(t *testing.T) {
	// 1. กำหนดเวลาคงที่ (Fixed Time) ไว้ที่ด้านบนสุด
	fixedTime := time.Date(2026, time.June, 19, 16, 0, 0, 0, time.Local)
	mockInput := dto.UpdateTransactionInput{
		Amount:          pkg.PTR(200.0),
		Note:            pkg.PTR("edit amount"),
		CategoryID:      pkg.PTR(int64(3)),
		TransactionDate: pkg.PTR(fixedTime),
	}
	mockFetchTransaction := &domain.Transaction{
		ID:         1,
		Amount:     0,
		Note:       "",
		CategoryID: int64(1),
	}

	mockResult := &domain.Transaction{
		ID:              1,
		Amount:          200,
		CategoryID:      3,
		Note:            "edit amount",
		TransactionDate: fixedTime,
	}

	tests := []struct {
		name           string
		id             uint
		input          dto.UpdateTransactionInput
		setupMock      func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock)
		expectedResult *domain.Transaction
		expectedError  bool
	}{
		{
			name:  "1. Database Filure - ไม่สามารถหาข้อมูล​ Transaction จาก ID ได้",
			id:    uint(99),
			input: mockInput,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(99)).Return(nil, domain.ErrNotFound)

			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:  "2. Sucess case - แก้ไขข้อมูลสำเร็จ",
			id:    uint(1),
			input: mockInput,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(mockFetchTransaction, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(tx *domain.Transaction) bool {
					return tx.ID == 1 &&
						tx.Amount == 200 &&
						tx.Note == "edit amount" &&
						tx.CategoryID == 3

				})).Return(nil)

				cache.On("InvalidateCache", mock.Anything, mock.Anything).Return(nil)
			},
			expectedResult: mockResult,
			expectedError:  false,
		},
		{
			name:  "3. Datebase Error - ฐานข้อมูลมีปัญหาไม่สามารถระบุได้",
			id:    uint(1),
			input: mockInput,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(mockFetchTransaction, nil)
				repo.On("Update", mock.Anything, mockFetchTransaction).Return(domain.ErrInternalDB)

			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(domain.TransactionRepositoryMock)
			mockCache := new(domain.TransactionCacheRepositoryMock)
			mockGemini := new(domain.GeminiSlipRepositoryMock)
			mockLogger := logger.NewNopLogger()

			ctx := context.Background()

			tt.setupMock(mockRepo, mockCache)

			uc := usecase.NewTransactionUsecase(mockRepo, mockCache, mockGemini, mockLogger)

			// Act
			result, err := uc.UpdateTransaction(ctx, tt.id, tt.input)

			// Assert
			if tt.expectedError {
				assert.Nil(t, result)
				assert.NotNil(t, err)
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.ID, tt.id)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)

		})
	}
}

func TestSyncTransaction(t *testing.T) {
	ctx := context.Background()
	fakeBytes := []byte("fake-image-data")
	aiDate := "2006-01-02 15:04:05"
	mockTime, _ := timeutil.ParseAISlipTime(aiDate)

	tests := []struct {
		name           string
		imageBytes     []byte
		localImageName string
		setupMock      func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, gemini *domain.GeminiSlipRepositoryMock)
		expectedAssert func(t *testing.T, result *domain.Transaction, err error)
	}{
		{
			name:           "1. Early Short-Circuit - ไฟล์เคยประมวลผลแล้ว",
			imageBytes:     fakeBytes,
			localImageName: "already_done.jpg",
			setupMock: func(
				repo *domain.TransactionRepositoryMock,
				cache *domain.TransactionCacheRepositoryMock,
				gemini *domain.GeminiSlipRepositoryMock,
			) {
				// จำลองว่า CheckFileExists เจอไฟล์นี้แล้ว (true,nil)
				cache.On("CheckFileExists", ctx, "already_done.jpg").Return(true, nil)

				// ส่วนอื่นจะไม่ได้ทำงานทันที
			},
			expectedAssert: func(t *testing.T, result *domain.Transaction, err error) {
				assert.NoError(t, err)
				assert.Nil(t, result)
			},
		},
		{
			name:           "2. Gemini Extraction Failed - ส่งหา AI แล้วพัง",
			imageBytes:     fakeBytes,
			localImageName: "new_slip.jpg",
			setupMock: func(
				repo *domain.TransactionRepositoryMock,
				cache *domain.TransactionCacheRepositoryMock,
				gemini *domain.GeminiSlipRepositoryMock,
			) {
				cache.On("CheckFileExists", ctx, "new_slip.jpg").Return(false, nil)

				// จำลองว่า AI คินค่า Error
				gemini.On("ExtractData", ctx, fakeBytes).Return(nil, errors.New("gemini api error"))

			},
			expectedAssert: func(t *testing.T, result *domain.Transaction, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "gemini api error")
				assert.Nil(t, result)
			},
		},
		{
			name:           "3. Success Path - ข้อมูลใหม่ทั้งหมด (ดักจับ Amount ติดลบด้วย)",
			imageBytes:     fakeBytes,
			localImageName: "perfact_slip.jpg",
			setupMock: func(
				repo *domain.TransactionRepositoryMock,
				cache *domain.TransactionCacheRepositoryMock,
				gemini *domain.GeminiSlipRepositoryMock,
			) {
				cache.On("CheckFileExists", ctx, "perfact_slip.jpg").Return(false, nil)

				mockSlipData := &domain.GeminiSlipData{
					Amount:       200,
					SenderName:   "ja",
					ReceiverName: "Gash mach",
					TransTime:    aiDate,
				}
				gemini.On("ExtractData", ctx, fakeBytes).Return(mockSlipData, nil)

				// ตรวจสอบข้อมูลที่จะลงฐานข้อมูล
				expectedTx := &domain.Transaction{
					Amount:          200,
					TransactionType: "EXPENSE",
					ReceiverName:    "Gash mach",
					LocalImageName:  "perfact_slip.jpg",
					TransactionDate: mockTime,
				}
				repo.On("Insert", ctx, expectedTx).Return(nil)

				// หลังเซฟเสร็จ ต้องทำ 2 อย่าง: เซ็ตไฟล์แคช และ ล้างแคชแดชบอร์ด
				cache.On("SetFileCache", ctx, "perfact_slip.jpg").Return(nil)

				periodKey := fmt.Sprintf("summary:monthly:%d-%02d", expectedTx.TransactionDate.Year(), expectedTx.TransactionDate.Month())
				cache.On("InvalidateCache", ctx, periodKey).Return(nil)
			},
			expectedAssert: func(t *testing.T, result *domain.Transaction, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, float64(200), result.Amount)
				assert.Equal(t, "EXPENSE", result.TransactionType)
			},
		},
		{
			name:           "4. Failed Insert to DB - ฐานข้อมูลมีปัญหา",
			imageBytes:     fakeBytes,
			localImageName: "perfect_slip.jpg",
			setupMock: func(
				repo *domain.TransactionRepositoryMock,
				cache *domain.TransactionCacheRepositoryMock,
				gemini *domain.GeminiSlipRepositoryMock,
			) {
				cache.On("CheckFileExists", ctx, "perfect_slip.jpg").Return(false, nil)

				mockSlipData := &domain.GeminiSlipData{
					Amount:       200,
					SenderName:   "ja",
					ReceiverName: "Gash mach",
					TransTime:    aiDate,
				}
				gemini.On("ExtractData", ctx, fakeBytes).Return(mockSlipData, nil)

				// ตรวจสอบข้อมูลที่จะลงฐานข้อมูล (ต้องเปลี่ยนจาก -50.00 เป็น 0.00)
				expectedTx := &domain.Transaction{
					Amount:          200,
					TransactionType: "EXPENSE",
					ReceiverName:    "Gash mach",
					LocalImageName:  "perfect_slip.jpg",
					TransactionDate: mockTime,
				}
				repo.On("Insert", ctx, expectedTx).Return(errors.New("db error"))

			},
			expectedAssert: func(t *testing.T, result *domain.Transaction, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attange
			mockRepo := new(domain.TransactionRepositoryMock)
			mockCache := new(domain.TransactionCacheRepositoryMock)
			mockGemini := new(domain.GeminiSlipRepositoryMock)
			mockLogger := logger.NewNopLogger()

			tt.setupMock(mockRepo, mockCache, mockGemini)

			txUsecase := usecase.NewTransactionUsecase(mockRepo, mockCache, mockGemini, mockLogger)

			// Act
			result, err := txUsecase.SyncTransaction(ctx, tt.imageBytes, tt.localImageName)

			// Assert
			tt.expectedAssert(t, result, err)

			// Verify
			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
			mockGemini.AssertExpectations(t)
		})
	}
}
