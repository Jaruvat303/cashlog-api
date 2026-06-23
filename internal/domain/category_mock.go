package domain

import (
	"context"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/stretchr/testify/mock"
)

// CategoryRepositoryMock
type CategoryRepositoryMock struct {
	mock.Mock
}

func (m *CategoryRepositoryMock) Create(ctx context.Context, category *Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *CategoryRepositoryMock) Update(ctx context.Context, updateCat *Category, id uint) error {
	args := m.Called(ctx, updateCat, id)
	return args.Error(0)
}

func (m *CategoryRepositoryMock) GetAll(ctx context.Context) ([]Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Category), args.Error(1)
}

func (m *CategoryRepositoryMock) GetByID(ctx context.Context, id uint) (*Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}

func (m *CategoryRepositoryMock) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// CategoryUsecaseMock
type CategoryUsecaseMock struct {
	mock.Mock
}

func (m *CategoryUsecaseMock) CreateCategory(ctx context.Context, input dto.CreateCategoryInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
func (m *CategoryUsecaseMock) GetCategoryByID(ctx context.Context, id uint) (*Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}
func (m *CategoryUsecaseMock) FetchCategories(ctx context.Context) ([]Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Category), args.Error(1)
}
func (m *CategoryUsecaseMock) UpdateCategory(ctx context.Context, id uint, input dto.UpdateCategoryInput) (*Category, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}
func (m *CategoryUsecaseMock) DeleteCategory(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
