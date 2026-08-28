package usecase

import (
	"context"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
)

type accountUsecase struct {
	accountRepo domain.AccountRepo
	log         logger.Logger
}

// CreateAccount implements [domain.AccountUsecase].
func (a *accountUsecase) CreateAccount(ctx context.Context, input domain.CreateAccountParam) (*domain.Account, error) {
	acc := &domain.Account{
		Name:             input.Name,
		AccountType:      input.AccountType,
		OpeningBalance:   input.OpeningBalance,
		MatchingKeywords: domain.StringSlice(input.MatchingKeywords),
		IconKey:          input.IconKey,
		ColorHex:         input.ColorHex,
	}

	if err := a.accountRepo.Create(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

// DeleteAccount implements [domain.AccountUsecase].
func (a *accountUsecase) DeleteAccount(ctx context.Context, id uint) error {
	acc, err := a.accountRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if acc != nil {
		if err := a.accountRepo.Delete(ctx, id); err != nil {
			return err
		}
	}

	return nil
}

// FetchActiveAccounts implements [domain.AccountUsecase].
func (a *accountUsecase) FetchActiveAccounts(ctx context.Context) ([]domain.Account, error) {
	return a.accountRepo.GetAllActive(ctx)
}

// UpdateAccount implements [domain.AccountUsecase].
func (a *accountUsecase) UpdateAccount(ctx context.Context, id uint, input domain.UpdateAccountParam) (*domain.Account, error) {
	acc, err := a.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		acc.Name = input.Name
	}
	if input.AccountType != "" {
		acc.AccountType = input.AccountType
	}
	if input.MatchingKeywords != nil {
		acc.MatchingKeywords = domain.StringSlice(input.MatchingKeywords)
	}
	if input.IconKey != "" {
		acc.IconKey = input.IconKey
	}
	if input.ColorHex != "" {
		acc.ColorHex = input.ColorHex
	}
	if input.IsActive != nil {
		acc.IsActive = *input.IsActive
	}

	if err := a.accountRepo.Update(ctx, acc, id); err != nil {
		return nil, err
	}

	return acc, nil
}

func NewAccountUsecase(accountRepo domain.AccountRepo, appLogger logger.Logger) domain.AccountUsecase {
	return &accountUsecase{
		accountRepo: accountRepo,
		log:         appLogger,
	}
}
