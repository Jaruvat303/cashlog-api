package usecase_test

import (
	"context"
	"testing"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/internal/usecase"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAccount(t *testing.T) {
	mockInput := domain.CreateAccountParam{
		Name:             "test account",
		AccountType:      "cash",
		OpeningBalance:   1000.00,
		MatchingKeywords: []string{"test", "account"},
		IconKey:          "icon_test",
		ColorHex:         "#FFFFFF",
	}

	mockAcc := &domain.Account{
		Name:             mockInput.Name,
		AccountType:      mockInput.AccountType,
		OpeningBalance:   mockInput.OpeningBalance,
		MatchingKeywords: domain.StringSlice(mockInput.MatchingKeywords),
		IconKey:          mockInput.IconKey,
		ColorHex:         mockInput.ColorHex,
	}

	tests := []struct {
		name          string
		input         domain.CreateAccountParam
		setupMock     func(repo *domain.AccountRepositoryMock)
		expectedError error
	}{
		{
			name:  "1. Success - สร้างข้อมูล Account สำเร็จ",
			input: mockInput,
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("Create", mock.Anything, mockAcc).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "2. DB Error - ฐานข้อมูลขัดข้อง",
			input: mockInput,
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("Create", mock.Anything, mockAcc).Return(domain.ErrInternalDB)
			},
			expectedError: domain.ErrInternalDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arange
			mockRepo := new(domain.AccountRepositoryMock)
			mockLog := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo)

			// Act
			accountUsecase := usecase.NewAccountUsecase(mockRepo, mockLog)
			account, err := accountUsecase.CreateAccount(ctx, tt.input)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, account)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	mockAcc := &domain.Account{
		ID:               1,
		Name:             "test account",
		AccountType:      "cash",
		OpeningBalance:   1000.00,
		MatchingKeywords: domain.StringSlice([]string{"test", "account"}),
		IconKey:          "icon_test",
		ColorHex:         "#FFFFFF",
	}

	tests := []struct {
		name          string
		accountID     uint
		setupMock     func(repo *domain.AccountRepositoryMock)
		expectedError error
	}{
		{
			name:      "1. Success - ลบข้อมูล Account สำเร็จ",
			accountID: uint(mockAcc.ID),
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(mockAcc.ID)).Return(mockAcc, nil)
				repo.On("Delete", mock.Anything, uint(mockAcc.ID)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "2. Not Found - ไม่พบข้อมูล Account",
			accountID: uint(mockAcc.ID),
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(mockAcc.ID)).Return(nil, nil)
			},
			expectedError: nil,
		},
		{
			name:      "3. DB Error - ฐานข้อมูลขัดข้อง",
			accountID: uint(mockAcc.ID),
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(mockAcc.ID)).Return(nil, domain.ErrInternalDB)
			},
			expectedError: domain.ErrInternalDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arange
			mockRepo := new(domain.AccountRepositoryMock)
			mockLog := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo)

			// Act
			accountUsecase := usecase.NewAccountUsecase(mockRepo, mockLog)
			err := accountUsecase.DeleteAccount(ctx, tt.accountID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFetchActiveAccounts(t *testing.T) {
	mockAcc1 := &domain.Account{
		ID:               1,
		Name:             "test account 1",
		AccountType:      "cash",
		OpeningBalance:   1000.00,
		MatchingKeywords: domain.StringSlice([]string{"test", "account"}),
		IconKey:          "icon_test_1",
		ColorHex:         "#FFFFFF",
		IsActive:         true,
	}

	mockAcc2 := &domain.Account{
		ID:               2,
		Name:             "test account 2",
		AccountType:      "bank",
		OpeningBalance:   2000.00,
		MatchingKeywords: domain.StringSlice([]string{"test", "account"}),
		IconKey:          "icon_test_2",
		ColorHex:         "#000000",
		IsActive:         true,
	}

	tests := []struct {
		name           string
		setupMock      func(repo *domain.AccountRepositoryMock)
		expectedError  error
		expectedResult []domain.Account
	}{
		{
			name: "1. Success - ดึงข้อมูล Account ที่ Active สำเร็จ",
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("GetAllActive", mock.Anything).Return([]domain.Account{*mockAcc1, *mockAcc2}, nil)
			},
			expectedError:  nil,
			expectedResult: []domain.Account{*mockAcc1, *mockAcc2},
		},
		{
			name: "2. DB Error - ฐานข้อมูลขัดข้อง",
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("GetAllActive", mock.Anything).Return(nil, domain.ErrInternalDB)
			},
			expectedError:  domain.ErrInternalDB,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arange
			mockRepo := new(domain.AccountRepositoryMock)
			mockLog := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo)

			// Act
			accountUsecase := usecase.NewAccountUsecase(mockRepo, mockLog)
			result, err := accountUsecase.FetchActiveAccounts(ctx)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	mockInput := domain.UpdateAccountParam{
		Name:             "updated account",
		AccountType:      "bank",
		MatchingKeywords: []string{"updated", "account"},
		IconKey:          "icon_updated",
		ColorHex:         "#000000",
		IsActive:         new(bool),
	}

	mockAcc := &domain.Account{
		ID:               1,
		Name:             "test account",
		AccountType:      "cash",
		OpeningBalance:   1000.00,
		MatchingKeywords: domain.StringSlice([]string{"test", "account"}),
		IconKey:          "icon_test",
		ColorHex:         "#FFFFFF",
		IsActive:         true,
	}

	tests := []struct {
		name          string
		accountID     uint
		input         domain.UpdateAccountParam
		setupMock     func(repo *domain.AccountRepositoryMock)
		expectedError error
	}{
		{
			name:      "1. Success - แก้ไขข้อมูล Account สำเร็จ",
			accountID: uint(mockAcc.ID),
			input:     mockInput,
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(mockAcc.ID)).Return(mockAcc, nil)
				repo.On("Update", mock.Anything, mockAcc, uint(mockAcc.ID)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "2. Not Found - ไม่พบข้อมูล Account",
			accountID: uint(mockAcc.ID),
			input:     mockInput,
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(mockAcc.ID)).Return(nil, domain.ErrNotFound)
			},
			expectedError: domain.ErrNotFound,
		},
		{
			name:      "3. DB Error - ฐานข้อมูลขัดข้อง",
			accountID: uint(mockAcc.ID),
			input:     mockInput,
			setupMock: func(repo *domain.AccountRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(mockAcc.ID)).Return(nil, domain.ErrInternalDB)
			},
			expectedError: domain.ErrInternalDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arange
			mockRepo := new(domain.AccountRepositoryMock)
			mockLog := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo)

			// Act
			accountUsecase := usecase.NewAccountUsecase(mockRepo, mockLog)
			result, err := accountUsecase.UpdateAccount(ctx, tt.accountID, tt.input)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
