package dto

import "github.com/Jaruvat303/cashlog/internal/domain"

// AccountResponse - ข้อมูลส่งออกกลับไปให้ Client (Response Body)
type AccountResponse struct {
	ID               int64    `json:"id" example:"1"`
	Name             string   `json:"name" example:"SCB"`
	AccountType      string   `json:"account_type" example:"bank"`
	OpeningBalance   float64  `json:"opening_balance" example:"1000"`
	MatchingKeywords []string `json:"matching_keywords" example:"SCB"`
	IconKey          string   `json:"icon_key" example:"bank"`
	ColorHex         string   `json:"color_hex" example:"#4F46E5"`
	IsActive         bool     `json:"is_active" example:"true"`
}

// CreateAccountInput - ข้อมูลนำเข้าสำหรับสร้างบัญชี (Request Body)
type CreateAccountInput struct {
	Name             string   `json:"name" validate:"required,min=1,max=100" example:"SCB"`
	AccountType      string   `json:"account_type" validate:"required,oneof=cash bank investment ewallet" example:"bank"`
	OpeningBalance   float64  `json:"opening_balance" validate:"gte=0" example:"1000"`
	MatchingKeywords []string `json:"matching_keywords" validate:"omitempty,dive,required" example:"SCB"`
	IconKey          string   `json:"icon_key" validate:"omitempty,max=50" example:"bank"`
	ColorHex         string   `json:"color_hex" validate:"omitempty,max=10" example:"#4F46E5"`
}

// UpdateAccountInput - ข้อมูลนำเข้าสำหรับอัปเดตบัญชี (Request Body)
type UpdateAccountInput struct {
	Name             *string   `json:"name" validate:"omitempty,min=1,max=100" example:"SCB"`
	AccountType      *string   `json:"account_type" validate:"omitempty,oneof=cash bank investment ewallet" example:"bank"`
	MatchingKeywords *[]string `json:"matching_keywords" validate:"omitempty,dive,required" example:"SCB"`
	IconKey          *string   `json:"icon_key" validate:"omitempty,max=50" example:"bank"`
	ColorHex         *string   `json:"color_hex" validate:"omitempty,max=10" example:"#4F46E5"`
	IsActive         *bool     `json:"is_active" example:"true"`
}

// ToDomainCreateParam แปลงข้อมูลจาก DTO เป็น Domain Param
func (c *CreateAccountInput) ToDomainCreateParam() domain.CreateAccountParam {
	return domain.CreateAccountParam{
		Name:             c.Name,
		AccountType:      c.AccountType,
		OpeningBalance:   c.OpeningBalance,
		MatchingKeywords: c.MatchingKeywords,
		IconKey:          c.IconKey,
		ColorHex:         c.ColorHex,
	}
}

// ToDomainUpdateParam แปลงข้อมูลจาก DTO เป็น Domain Param
func (u *UpdateAccountInput) ToDomainUpdateParam() domain.UpdateAccountParam {
	param := domain.UpdateAccountParam{}

	if u.Name != nil {
		param.Name = *u.Name
	}

	if u.AccountType != nil {
		param.AccountType = *u.AccountType
	}

	if u.MatchingKeywords != nil {
		param.MatchingKeywords = *u.MatchingKeywords
	}

	if u.IconKey != nil {
		param.IconKey = *u.IconKey
	}

	if u.ColorHex != nil {
		param.ColorHex = *u.ColorHex
	}

	if u.IsActive != nil {
		param.IsActive = u.IsActive
	}

	return param
}

func MapToAccountResponse(a *domain.Account) AccountResponse {
	return AccountResponse{
		ID:               a.ID,
		Name:             a.Name,
		AccountType:      a.AccountType,
		OpeningBalance:   a.OpeningBalance,
		MatchingKeywords: []string(a.MatchingKeywords),
		IconKey:          a.IconKey,
		ColorHex:         a.ColorHex,
		IsActive:         a.IsActive,
	}
}

func MapToAccountListResponse(accounts []domain.Account) []AccountResponse {
	res := make([]AccountResponse, len(accounts))
	for i, acc := range accounts {
		res[i] = MapToAccountResponse(&acc)
	}
	return res
}
