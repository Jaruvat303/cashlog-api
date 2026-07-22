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

func TestCreateCategory(t *testing.T) {
	mockInput := domain.CreateCategoryParam{
		Name: "food",
		Type: "expense",
	}

	mockCat := &domain.Category{
		Name: mockInput.Name,
		Type: mockInput.Type,
	}

	tests := []struct {
		name          string
		input         domain.CreateCategoryParam
		setupMock     func(repo *domain.CategoryRepositoryMock)
		expectedError error
	}{
		{
			name: "1. Success - สร้างข้อมูล Category สำเร็จ",
			input: domain.CreateCategoryParam{
				Name: mockInput.Name,
				Type: mockInput.Type,
			},
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("Create", mock.Anything, mockCat).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "2. DB Error - ฐานข้อมูลขัดข้อง",
			input: mockInput,
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("Create", mock.Anything, mockCat).Return(domain.ErrInternalDB)
			},
			expectedError: domain.ErrInternalDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arange
			mockRepo := new(domain.CategoryRepositoryMock)
			mockLog := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo)

			uc := usecase.NewCategoryUsecase(mockRepo, mockLog)

			// Act
			result, err := uc.CreateCategory(ctx, tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	mockInput := domain.UpdateCategoryParam{
		Name: "update food",
	}

	mockCategpry := &domain.Category{
		ID:   1,
		Name: "food",
		Type: "expense",
	}

	mockResult := &domain.Category{
		ID:   1,
		Name: mockInput.Name,
		Type: mockCategpry.Type,
	}

	tests := []struct {
		name           string
		id             uint
		input          domain.UpdateCategoryParam
		setupMock      func(repo *domain.CategoryRepositoryMock)
		expectedResult *domain.Category
		expectedError  bool
	}{
		{
			name:  "1. Error Notfound - Category ID ไม่พบข้อมูลที่จะแก่ไข",
			id:    99,
			input: domain.UpdateCategoryParam{},
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(99)).Return(nil, domain.ErrNotFound)
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:  "2. Success - Update Category Complete",
			id:    1,
			input: mockInput,
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(mockCategpry, nil)
				repo.On("Update", mock.Anything, mockResult, uint(1)).Return(nil)
			},
			expectedResult: mockResult,
			expectedError:  false,
		},
		{
			name:  "3. DB Error - Update Category Failed",
			id:    1,
			input: mockInput,
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(mockCategpry, nil)
				repo.On("Update", mock.Anything, mockResult, uint(1)).Return(domain.ErrInternalDB)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(domain.CategoryRepositoryMock)
			mockLogger := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo)

			uc := usecase.NewCategoryUsecase(mockRepo, mockLogger)

			// Act
			result, err := uc.UpdateCategory(ctx, tt.id, tt.input)

			// Assert
			if tt.expectedError {
				assert.Nil(t, result)
				assert.Error(t, err)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.ID, result.ID)
				assert.Equal(t, tt.expectedResult.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFetchCategoriesByType(t *testing.T) {
	mockCategories := []domain.Category{
		{ID: 1, Name: "food"}, {ID: 2, Name: "healty", Type: "expense"},
	}

	tests := []struct {
		name           string
		types          string
		setupMock      func(repo *domain.CategoryRepositoryMock)
		expectedResult []domain.Category
		expectedError  bool
	}{
		{
			name:  "1. Success - Get All Categories",
			types: "expense",
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("GetByType", mock.Anything, "expense").Return(mockCategories, nil)
			},
			expectedResult: mockCategories,
			expectedError:  false,
		},
		{
			name:  "2. DB Error - Cannot Get All Categories",
			types: "expense",
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("GetByType", mock.Anything, "expense").Return(nil, domain.ErrInternalDB)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(domain.CategoryRepositoryMock)
			mockLog := logger.NewNopLogger()

			tt.setupMock(mockRepo)

			ctx := context.Background()

			uc := usecase.NewCategoryUsecase(mockRepo, mockLog)

			// Act
			result, err := uc.FetchCategoriesByType(ctx, tt.types)

			// Assart
			if tt.expectedError {
				assert.Nil(t, result)
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	mockCat := &domain.Category{
		ID:   1,
		Name: "food",
	}

	tests := []struct {
		name          string
		id            uint
		setupMock     func(repo *domain.CategoryRepositoryMock)
		expectedError error
	}{
		{
			name: "1. Category Not Found  - can not get category from data database",
			id:   99,
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(99)).Return(nil, domain.ErrNotFound)
			},
			expectedError: domain.ErrNotFound,
		},
		{
			name: "2. DB Internal Error - can not delete category from database",
			id:   1,
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(mockCat, nil)
				repo.On("Delete", mock.Anything, uint(1)).Return(domain.ErrInternalDB)
			},
			expectedError: domain.ErrInternalDB,
		},
		{
			name: "3. Delete Category Successfully",
			id:   1,
			setupMock: func(repo *domain.CategoryRepositoryMock) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(mockCat, nil)
				repo.On("Delete", mock.Anything, uint(1)).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(domain.CategoryRepositoryMock)
			mockLog := logger.NewNopLogger()
			ctx := context.Background()

			tt.setupMock(mockRepo)

			uc := usecase.NewCategoryUsecase(mockRepo, mockLog)

			// Act
			err := uc.DeleteCategory(ctx, tt.id)

			// Assert
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.Nil(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
