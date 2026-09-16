package usecase

import (
	"context"
	"regexp"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
)

// defaultIconKey / defaultColorHex คือค่า icon/color เดียวกับ category "ไม่ระบุประเภท" ที่ seed ไว้แล้ว
const (
	defaultIconKey  = "question-fill"
	defaultColorHex = "#64748B"
)

// colorHexPattern ต้องเป็น # ตามด้วยเลขฐาน 16 จำนวน 6 หลัก เช่น "#64748B"
var colorHexPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type categoryUsecase struct {
	categoryRepo domain.CategoryRepo
	log          logger.Logger
}

// CreateCategory implements [domain.CategoryUsecase].
func (c *categoryUsecase) CreateCategory(ctx context.Context, input domain.CreateCategoryParam) (*domain.Category, error) {
	iconKey := input.IconKey
	if iconKey == "" {
		iconKey = defaultIconKey
	}

	colorHex := input.ColorHex
	if colorHex == "" {
		colorHex = defaultColorHex
	} else if !colorHexPattern.MatchString(colorHex) {
		return nil, domain.ErrInvalidColorHex
	}

	cat := &domain.Category{
		Name:     input.Name,
		Type:     input.Type,
		IconKey:  iconKey,
		ColorHex: colorHex,
	}

	if err := c.categoryRepo.Create(ctx, cat); err != nil {
		return nil, err
	}

	return cat, nil
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
func (c *categoryUsecase) FetchCategoriesByType(ctx context.Context, types string) ([]domain.Category, error) {
	return c.categoryRepo.GetByType(ctx, types)
}

// UpdateCategory implements [domain.CategoryUsecase].
func (c *categoryUsecase) UpdateCategory(ctx context.Context, id uint, input domain.UpdateCategoryParam) (*domain.Category, error) {
	if input.ColorHex != "" && !colorHexPattern.MatchString(input.ColorHex) {
		return nil, domain.ErrInvalidColorHex
	}

	// ค้นหาข้อมูล category ในฐานข้อมูล
	cat, err := c.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		cat.Name = input.Name
	}

	if input.IconKey != "" {
		cat.IconKey = input.IconKey
	}

	if input.ColorHex != "" {
		cat.ColorHex = input.ColorHex
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
