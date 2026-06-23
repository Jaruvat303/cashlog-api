package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
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
		setupMock      func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock)
		expectedResult []domain.Transaction
		expectedError  bool
	}{
		{
			name:  "1. Successfully",
			month: 6,
			year:  2026,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
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
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
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
			mockOCR := new(domain.OCRGatewayMock)
			mockLogger := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo, mockCacheRepo, mockOCR)

			txUsecase := usecase.NewTransactionUsecase(mockRepo, mockCacheRepo, mockOCR, mockLogger)

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
			mockOCR.AssertExpectations(t)
		})
	}
}

func TestGetDashboardSummary(t *testing.T) {
	// จำลองข่อมูลที่จะส่งกลับมาจาก Cache หรือ DB
	mockSummary := &domain.DashboardSummary{
		TotalIncome:  15000.0,
		TotalExpence: 5000.0,
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
		setupMock      func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock)
		expectedResult *domain.DashboardSummary
		expectedError  bool
	}{
		{
			name:  "1. Cache Hit - Monthly",
			scope: "monthly",
			month: 6,
			year:  2026,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
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
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
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
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
				ctx := mock.Anything
				periodKey := "summary:year:2026"

				// กำหนดวันเริ่มต้นและวันสุดท้าย
				startDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, timeutil.BangKokLoc)
				endDate := startDate.AddDate(1, 0, 0).Add(-time.Nanosecond)

				yearlySummary := &domain.DashboardSummary{
					TotalIncome:  150000.0,
					TotalExpence: 40000.0,
					Scope:        "yearly",
					Year:         2026,
				}

				cache.On("GetCache", ctx, periodKey).Return("", errors.New("cache miss"))
				repo.On("CalculateSummary", ctx, startDate, endDate, "yearly").Return(yearlySummary, nil)
				cache.On("SetCache", ctx, periodKey, mock.Anything).Return(nil)
			},
			expectedResult: &domain.DashboardSummary{
				TotalIncome:  150000.0,
				TotalExpence: 40000.0,
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
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
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
			mockOCR := new(domain.OCRGatewayMock)
			mockLogger := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo, mockCache, mockOCR)

			logger.InitLogger("development")
			// นำ Mock ไปใส่ใน usecase
			txUsecase := usecase.NewTransactionUsecase(mockRepo, mockCache, mockOCR, mockLogger)

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
				assert.Equal(t, tt.expectedResult.TotalExpence, result.TotalExpence)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
			mockOCR.AssertExpectations(t)

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
		setupMock     func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock)
		expectedError bool
	}{
		{
			name: "1. DB Error - ค้นหาข้อมูล Transaction ไม่พบ",
			id:   99,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
				repo.On("GetByID", mock.Anything, uint(99)).Return(nil, domain.ErrNotFound)
			},

			expectedError: true,
		},
		{
			name: "2. DB Error - ลบข้อมูลไม่สำเร็จ ฐานข้อมูลมีปัญหา",
			id:   1,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(mockFetchTransaction, nil)
				repo.On("Delete", mock.Anything, uint(1)).Return(domain.ErrInternalDB)
			},
			expectedError: true,
		},
		{
			name: "3. Success - ลบข้อมูลสำเร็จ",
			id:   1,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
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
			mockOCR := new(domain.OCRGatewayMock)
			mockLogger := logger.NewNopLogger()

			tt.setupMock(mockRepo, mockCache, mockOCR)

			ctx := context.Background()

			uc := usecase.NewTransactionUsecase(mockRepo, mockCache, mockOCR, mockLogger)

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
			mockOCR.AssertExpectations(t)

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
		setupMock      func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock)
		expectedResult *domain.Transaction
		expectedError  bool
	}{
		{
			name:  "1. Database Filure - ไม่สามารถหาข้อมูล​ Transaction จาก ID ได้",
			id:    uint(99),
			input: mockInput,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
				repo.On("GetByID", mock.Anything, uint(99)).Return(nil, domain.ErrNotFound)

			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:  "2. Sucess case - แก้ไขข้อมูลสำเร็จ",
			id:    uint(1),
			input: mockInput,
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
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
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
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
			mockOCR := new(domain.OCRGatewayMock)
			mockLogger := logger.NewNopLogger()

			ctx := context.Background()

			tt.setupMock(mockRepo, mockCache, mockOCR)

			uc := usecase.NewTransactionUsecase(mockRepo, mockCache, mockOCR, mockLogger)

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
			mockOCR.AssertExpectations(t)

		})
	}
}

func TestSyncTransaction(t *testing.T) {
	ctx := context.Background()
	fakeBytes := []byte("fake-image-data")
	now := timeutil.NowInBangkok()

	tests := []struct {
		name           string
		imageBytes     []byte
		localImageName string
		setupMock      func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock)
		expectedAssert func(t *testing.T, result *domain.Transaction, err error)
	}{
		{
			name:           "1. Early Short-Circuit - ไฟล์เคยประมวลผลแล้ว",
			imageBytes:     fakeBytes,
			localImageName: "already_done.jpg",
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
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
			name:           "2. OCR Extraction Failed - ส่งหา AI แล้วพัง",
			imageBytes:     fakeBytes,
			localImageName: "new_slip.jpg",
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
				cache.On("CheckFileExists", ctx, "new_slip.jpg").Return(false, nil)

				// จำลองว่า AI คินค่า Error
				ocr.On("Extract", ctx, fakeBytes).Return(nil, errors.New("google vision api error"))

			},
			expectedAssert: func(t *testing.T, result *domain.Transaction, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "ocr extraction failed")
				assert.Nil(t, result)
			},
		},
		{
			name:           "3. Duplicate Transaction ID - รหัสธุกรรมซ้ำใน DB",
			imageBytes:     fakeBytes,
			localImageName: "new_file_but_dup_tx.jpg",
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
				cache.On("CheckFileExists", ctx, "new_file_but_dup_tx.jpg").Return(false, nil)

				mockORCData := &domain.OCRData{
					TransactionID:   "TX_DUP_1234",
					Amount:          250.00,
					TransactionDate: now,
				}
				ocr.On("Extract", ctx, fakeBytes).Return(mockORCData, nil)

				// จำลองว่าระบบไปเช็คกับ DB แล้วพบว่ารหัสธุรกรรมนี้เคยเซฟไปแล้ว (true, nil)
				repo.On("CheckDuplicate", ctx, "TX_DUP_1234").Return(true, nil)

				// ระบบต้องเอาชื่อใหม่ ไปบันทึกใน redis เพื่อป้องกันการบันทึกซ่ำรอบหน้า
				cache.On("SetFileCache", ctx, "new_file_but_dup_tx.jpg").Return(nil)
			},
			expectedAssert: func(t *testing.T, result *domain.Transaction, err error) {
				assert.NoError(t, err)
				assert.Nil(t, result)
			},
		},
		{
			name:           "4. Success Path - ข้อมูลใหม่ทั้งหมด (ดักจับ Amount ติดลบด้วย)",
			imageBytes:     fakeBytes,
			localImageName: "perfact_slip.jpg",
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
				cache.On("CheckFileExists", ctx, "perfact_slip.jpg").Return(false, nil)

				// จำลองเคส OCR พ้น Amount ติดลบ (-50) เพื่อทดสอบตัว Fallback
				mockOCRData := &domain.OCRData{
					TransactionID:   "TX_NEW_999",
					Amount:          -50,
					ReceiverName:    "ร้านค้าทดสอบ",
					TransactionDate: now,
				}
				ocr.On("Extract", ctx, fakeBytes).Return(mockOCRData, nil)
				repo.On("CheckDuplicate", ctx, "TX_NEW_999").Return(false, nil)

				// ตรวจสอบข้อมูลที่จะลงฐานข้อมูล (ต้องเปลี่ยนจาก -50.00 เป็น 0.00)
				expectedTx := &domain.Transaction{
					TransactionID:   "TX_NEW_999",
					Amount:          0.00,
					TransactionType: "EXPENSE",
					ReceiverName:    "ร้านค้าทดสอบ",
					LocalImageName:  "perfact_slip.jpg",
					TransactionDate: now,
				}
				repo.On("Insert", ctx, expectedTx).Return(nil)

				// หลังเซฟเสร็จ ต้องทำ 2 อย่าง: เซ็ตไฟล์แคช และ ล้างแคชแดชบอร์ด
				cache.On("SetFileCache", ctx, "perfact_slip.jpg").Return(nil)

				periodKey := now.Format("summary:monthly:2006-01") // สร้าง Format "summary:monthly:YYYY-MM" ตามโค้ดจริง
				cache.On("InvalidateCache", ctx, periodKey).Return(nil)
			},
			expectedAssert: func(t *testing.T, result *domain.Transaction, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, 0.0, result.Amount)
				assert.Equal(t, "EXPENSE", result.TransactionType)
				assert.Equal(t, "TX_NEW_999", result.TransactionID)
			},
		},
		{
			name:           "5. Failed Insert to DB - ฐานข้อมูลมีปัญหา",
			imageBytes:     fakeBytes,
			localImageName: "perfect_slip.jpg",
			setupMock: func(repo *domain.TransactionRepositoryMock, cache *domain.TransactionCacheRepositoryMock, ocr *domain.OCRGatewayMock) {
				cache.On("CheckFileExists", ctx, "perfect_slip.jpg").Return(false, nil)

				// จำลองเคส OCR พ่น Amount ติดลบมา (-50.00) เพื่อทดสอบตัว Fallback ลอจิกในโค้ดคุณด้วย
				mockOCRData := &domain.OCRData{
					TransactionID:   "TX_NEW_999",
					Amount:          -50.00,
					ReceiverName:    "ร้านค้าทดสอบ",
					TransactionDate: now,
				}
				ocr.On("Extract", ctx, fakeBytes).Return(mockOCRData, nil)
				repo.On("CheckDuplicate", ctx, "TX_NEW_999").Return(false, nil)

				// ตรวจสอบข้อมูลที่จะลงฐานข้อมูล (ต้องเปลี่ยนจาก -50.00 เป็น 0.00 ตามลอจิกคุณ)
				expectedTx := &domain.Transaction{
					TransactionID:   "TX_NEW_999",
					Amount:          0.00, // โค้ดคุณเก่งมาก ดักเปลี่ยนติดลบให้เป็น 0.00 อัตโนมัติ
					TransactionType: "EXPENSE",
					ReceiverName:    "ร้านค้าทดสอบ",
					LocalImageName:  "perfect_slip.jpg",
					TransactionDate: now,
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
			mockOCR := new(domain.OCRGatewayMock)
			mockLogger := logger.NewNopLogger()

			tt.setupMock(mockRepo, mockCache, mockOCR)

			txUsecase := usecase.NewTransactionUsecase(mockRepo, mockCache, mockOCR, mockLogger)

			// Act
			result, err := txUsecase.SyncTransaction(ctx, tt.imageBytes, tt.localImageName)

			// Assert
			tt.expectedAssert(t, result, err)

			// Verify
			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
			mockOCR.AssertExpectations(t)
		})
	}
}
