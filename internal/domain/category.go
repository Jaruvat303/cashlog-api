package domain

import (
	"context"
	"time"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
)

type Category struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(100);not null;uniqueIndex"`
	Type      string    `gorm:"type:varchar(100);not null;index"`
	IconURL   *string   `gorm:"type:varchar(255);column:icon_url"`
	CreatedAt time.Time `gorm:"autoCreateTime;not null"`
	UpdateAt  time.Time `gorm:"autoUpdateTime:not null"`
	Delete    time.Time `gorm:"index"`
}

type CategoryRepo interface {
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, updateCat *Category, id uint) error
	GetAll(ctx context.Context) ([]Category, error)
	GetByID(ctx context.Context, id uint) (*Category, error)
	Delete(ctx context.Context, id uint) error
}

type CategoryUsecase interface {
	CreateCategory(ctx context.Context, input dto.CreateCategoryInput) error
	FetchCategories(ctx context.Context) ([]Category, error)
	UpdateCategory(ctx context.Context, id uint, input dto.UpdateCategoryInput) (*Category, error)
	DeleteCategory(ctx context.Context, id uint) error
}
