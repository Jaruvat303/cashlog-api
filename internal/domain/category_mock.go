package domain

import (
	"context"

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

func (m *CategoryRepositoryMock) GetByType(ctx context.Context, types string) ([]Category, error) {
	args := m.Called(ctx, types)
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

func (m *CategoryUsecaseMock) CreateCategory(ctx context.Context, input CreateCategoryParam) (*Category, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}
func (m *CategoryUsecaseMock) GetCategoryByID(ctx context.Context, id uint) (*Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Category), args.Error(1)
}
func (m *CategoryUsecaseMock) FetchCategoriesByType(ctx context.Context, types string) ([]Category, error) {
	args := m.Called(ctx, types)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Category), args.Error(1)
}
func (m *CategoryUsecaseMock) UpdateCategory(ctx context.Context, id uint, input UpdateCategoryParam) (*Category, error) {
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
