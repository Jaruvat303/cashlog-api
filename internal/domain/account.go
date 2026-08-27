package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringSlice เก็บ list ของ string ลง column แบบ jsonb (ใช้กับ MatchingKeywords)
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("failed to scan StringSlice: unsupported type %T", value)
	}
}

type Account struct {
	ID               int64       `gorm:"primaryKey;autoIncrement"`
	Name             string      `gorm:"type:varchar(100);not null"`
	AccountType      string      `gorm:"type:varchar(50);not null"` // cash, bank, investment, ewallet
	OpeningBalance   float64     `gorm:"type:numeric(12,2);not null;default:0"`
	MatchingKeywords StringSlice `gorm:"type:jsonb;column:matching_keywords"`
	IconKey          string      `gorm:"type:varchar(50);column:icon_key"`
	ColorHex         string      `gorm:"type:varchar(10);column:color_hex"`
	IsActive         bool        `gorm:"not null;default:true"`
	CreatedAt        time.Time   `gorm:"autoCreateTime;not null"`
	UpdatedAt        time.Time   `gorm:"autoUpdateTime;not null"`
}

type AccountRepo interface {
	Create(ctx context.Context, account *Account) error
	Update(ctx context.Context, updateAcc *Account, id uint) error
	GetByID(ctx context.Context, id uint) (*Account, error)
	GetAllActive(ctx context.Context) ([]Account, error)
	Delete(ctx context.Context, id uint) error
}

type AccountUsecase interface {
	CreateAccount(ctx context.Context, input CreateAccountParam) (*Account, error)
	FetchActiveAccounts(ctx context.Context) ([]Account, error)
	UpdateAccount(ctx context.Context, id uint, input UpdateAccountParam) (*Account, error)
	DeleteAccount(ctx context.Context, id uint) error
}

type CreateAccountParam struct {
	Name             string
	AccountType      string
	OpeningBalance   float64
	MatchingKeywords []string
	IconKey          string
	ColorHex         string
}

type UpdateAccountParam struct {
	Name             string
	AccountType      string
	MatchingKeywords []string
	IconKey          string
	ColorHex         string
	IsActive         *bool
}
