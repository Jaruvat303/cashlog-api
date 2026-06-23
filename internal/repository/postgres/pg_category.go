package postgres

import (
	"context"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db  *gorm.DB
	log logger.Logger
}

// Create implements [domain.CategoryRepo].
func (c *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	err := c.db.WithContext(ctx).Create(category).Error
	if err != nil {
		return HandlerDBError(ctx, err, c.log)
	}
	return nil
}

// Delete implements [domain.CategoryRepo].
func (c *categoryRepository) Delete(ctx context.Context, id uint) error {
	err := c.db.WithContext(ctx).Delete(&domain.Category{}, id).Error
	if err != nil {
		return HandlerDBError(ctx, err, c.log)
	}
	return nil
}

// GetAll implements [domain.CategoryRepo].
func (c *categoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	result := c.db.WithContext(ctx).Order("id asc").Find(&categories)
	if result.Error != nil {
		return nil, HandlerDBError(ctx, result.Error, c.log)
	}
	return categories, nil
}

// GetByID implements [domain.CategoryRepo].
func (c *categoryRepository) GetByID(ctx context.Context, id uint) (*domain.Category, error) {
	var category domain.Category
	result := c.db.WithContext(ctx).First(&category, id)
	if result.Error != nil {
		return nil, HandlerDBError(ctx, result.Error, c.log)
	}
	return &category, nil
}

// Update implements [domain.CategoryRepo].
func (c *categoryRepository) Update(ctx context.Context, updateCat *domain.Category, id uint) error {
	err := c.db.WithContext(ctx).Model(updateCat).Select("*").Updates(updateCat).Error
	if err != nil {
		return HandlerDBError(ctx, err, c.log)
	}
	return nil
}

func NewGORMCategoryRepository(db *gorm.DB, appLogger logger.Logger) domain.CategoryRepo {
	return &categoryRepository{
		db:  db,
		log: appLogger,
	}
}
