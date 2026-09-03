package domain

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// AccountRepositoryMock
type AccountRepositoryMock struct {
	mock.Mock
}

func (m *AccountRepositoryMock) Create(ctx context.Context, account *Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *AccountRepositoryMock) Update(ctx context.Context, updateAcc *Account, id uint) error {
	args := m.Called(ctx, updateAcc, id)
	return args.Error(0)
}

func (m *AccountRepositoryMock) GetByID(ctx context.Context, id uint) (*Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *AccountRepositoryMock) GetAllActive(ctx context.Context) ([]Account, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Account), args.Error(1)
}

func (m *AccountRepositoryMock) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// AccountUsecaseMock
type AccountUsecaseMock struct {
	mock.Mock
}

func (m *AccountUsecaseMock) CreateAccount(ctx context.Context, input CreateAccountParam) (*Account, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *AccountUsecaseMock) FetchActiveAccounts(ctx context.Context) ([]Account, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Account), args.Error(1)
}

func (m *AccountUsecaseMock) UpdateAccount(ctx context.Context, id uint, input UpdateAccountParam) (*Account, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *AccountUsecaseMock) DeleteAccount(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
