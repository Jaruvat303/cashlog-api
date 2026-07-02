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
	IconKey   string    `gorm:"type:varchar(50);column:icon_key"`  // เก็บชื่อกลาง เช่น "utensils", "zap", "wallet"
	ColorHex  string    `gorm:"type:varchar(10);column:color_hex"` // เก็บโค้ดสี เช่น "#EF4444"
	CreatedAt time.Time `gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:not null"`
}

type CategoryRepo interface {
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, updateCat *Category, id uint) error
	GetByType(ctx context.Context,types string) ([]Category, error)
	GetByID(ctx context.Context, id uint) (*Category, error)
	Delete(ctx context.Context, id uint) error
}

type CategoryUsecase interface {
	CreateCategory(ctx context.Context, input dto.CreateCategoryInput) error
	FetchCategoriesByType(ctx context.Context, types string) ([]Category, error)
	UpdateCategory(ctx context.Context, id uint, input dto.UpdateCategoryInput) (*Category, error)
	DeleteCategory(ctx context.Context, id uint) error
}
