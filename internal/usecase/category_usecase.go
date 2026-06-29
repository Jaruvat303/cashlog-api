package usecase

import (
	"context"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
)

type categoryUsecase struct {
	categoryRepo domain.CategoryRepo
	log          logger.Logger
}

// CreateCategory implements [domain.CategoryUsecase].
func (c *categoryUsecase) CreateCategory(ctx context.Context, input dto.CreateCategoryInput) error {
	cat := &domain.Category{
		Name:    input.Name,
		Type:    input.Type,
		IconURl: input.IconURL,
	}

	if err := c.categoryRepo.Create(ctx, cat); err != nil {
		return err
	}

	return nil
}

// DeleteCategory implements [domain.CategoryUsecase].
func (c *categoryUsecase) DeleteCategory(ctx context.Context, id uint) error {
	cat, err := c.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if cat != nil {
		if err := c.categoryRepo.Delete(ctx, id); err != nil {
			return err
		}
	}

	return nil
}

// FetchCategories implements [domain.CategoryUsecase].
func (c *categoryUsecase) FetchCategories(ctx context.Context) ([]domain.Category, error) {
	return c.categoryRepo.GetAll(ctx)
}

// UpdateCategory implements [domain.CategoryUsecase].
func (c *categoryUsecase) UpdateCategory(ctx context.Context, id uint, input dto.UpdateCategoryInput) (*domain.Category, error) {
	// ค้นหาข้อมูล category ในฐานข้อมูล
	cat, err := c.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		cat.Name = *input.Name
	}

	if input.IconURL != nil {
		cat.IconURl = input.IconURL
	}

	// สั่ง Update category จากฐานข้อมูล
	if err := c.categoryRepo.Update(ctx, cat, id); err != nil {
		return nil, err
	}

	return cat, nil

}

func NewCategoryUsecase(categoryRepo domain.CategoryRepo, appLogger logger.Logger) domain.CategoryUsecase {
	return &categoryUsecase{
		categoryRepo: categoryRepo,
		log:          appLogger,
	}
}
