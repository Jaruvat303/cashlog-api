package postgres

import (
	"context"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"gorm.io/gorm"
)

type accountRepository struct {
	db  *gorm.DB
	log logger.Logger
}

// Create implements [domain.AccountRepo].
func (a *accountRepository) Create(ctx context.Context, account *domain.Account) error {
	err := a.db.WithContext(ctx).Create(account).Error
	if err != nil {
		return HandlerDBError(ctx, err, a.log)
	}
	return nil
}

// Update implements [domain.AccountRepo].
func (a *accountRepository) Update(ctx context.Context, updateAcc *domain.Account, id uint) error {
	err := a.db.WithContext(ctx).Model(updateAcc).Select("*").Updates(updateAcc).Error
	if err != nil {
		return HandlerDBError(ctx, err, a.log)
	}
	return nil
}

// GetByID implements [domain.AccountRepo].
func (a *accountRepository) GetByID(ctx context.Context, id uint) (*domain.Account, error) {
	var account domain.Account
	result := a.db.WithContext(ctx).First(&account, id)
	if result.Error != nil {
		return nil, HandlerDBError(ctx, result.Error, a.log)
	}
	return &account, nil
}

// GetAllActive implements [domain.AccountRepo].
func (a *accountRepository) GetAllActive(ctx context.Context) ([]domain.Account, error) {
	var accounts []domain.Account
	result := a.db.WithContext(ctx).Where("is_active = ?", true).Order("id asc").Find(&accounts)
	if result.Error != nil {
		return nil, HandlerDBError(ctx, result.Error, a.log)
	}
	return accounts, nil
}

// Delete implements [domain.AccountRepo].
// Soft delete: ตั้ง is_active = false แทนการลบแถวจริง ตาม pattern ที่ระบุใน domain.Account.IsActive
func (a *accountRepository) Delete(ctx context.Context, id uint) error {
	err := a.db.WithContext(ctx).Model(&domain.Account{}).Where("id = ?", id).Update("is_active", false).Error
	if err != nil {
		return HandlerDBError(ctx, err, a.log)
	}
	return nil
}

func NewGORMAccountRepository(db *gorm.DB, appLogger logger.Logger) domain.AccountRepo {
	return &accountRepository{
		db:  db,
		log: appLogger,
	}
}
