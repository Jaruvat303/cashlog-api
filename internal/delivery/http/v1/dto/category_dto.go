package dto

import "github.com/Jaruvat303/cashlog/internal/domain"

// CategoryResponse - ข้อมูลส่งออกกลับไปให้ Client (Response Body)
type CategoryResponse struct {
	ID       int64  `json:"id" example:"1"`
	Name     string `json:"name" example:"อาหาร"`
	Type     string `json:"type" example:"expense"`
	IconKey  string `json:"icon_key" example:"food"`
	ColorHex string `json:"color_hex" example:"#FF0000"`
}

// CreateCategoryInput - ข้อมูลนำเข้าสำหรับสร้างหมวดหมู่ (Request Body)
type CreateCategoryInput struct {
	Name     string  `json:"name" validate:"required,min=2,max=50" example:"อาหาร"`
	Type     string  `json:"type" validate:"required,oneof=income expense" example:"expense"`
	IconKey  *string `json:"icon_key" example:"food"`
	ColorHex *string `json:"color_hex" example:"#FF0000"`
}

// UpdateCategoryInput - ข้อมูลนำเข้าสำหรับอัปเดตหมวดหมู่ (Request Body)
type UpdateCategoryInput struct {
	Name     *string `json:"name" validate:"omitempty,min=3,max=50" example:"อาหาร"`
	Type     *string `json:"type" validate:"omitempty,oneof=income expense" example:"expense"`
	IconKey  *string `json:"icon_key" example:"food"`
	ColorHex *string `json:"color_hex" example:"#FF0000"`
}

// ToDomainCreateParam แปลงข้อมูลจาก DTO เป็น Domain Param
func (c *CreateCategoryInput) ToDomainCreateParam() domain.CreateCategoryParam {
	param := domain.CreateCategoryParam{
		Name: c.Name,
		Type: c.Type,
	}

	if c.IconKey != nil {
		param.IconKey = *c.IconKey
	}

	if c.ColorHex != nil {
		param.ColorHex = *c.ColorHex
	}

	return param
}

// ToDomainUpdateParam แปลงข้อมูลจาก DTO เป็น Domain Param
func (u *UpdateCategoryInput) ToDomainUpdateParam() domain.UpdateCategoryParam {
	param := domain.UpdateCategoryParam{}

	if u.Name != nil {
		param.Name = *u.Name
	}

	if u.Type != nil {
		param.Type = *u.Type
	}

	if u.IconKey != nil {
		param.IconKey = *u.IconKey
	}

	if u.ColorHex != nil {
		param.ColorHex = *u.ColorHex
	}

	return param
}

func MapToCategoryResponse(c *domain.Category) CategoryResponse {
	return CategoryResponse{
		ID:       c.ID,
		Name:     c.Name,
		Type:     c.Type,
		IconKey:  c.IconKey,
		ColorHex: c.ColorHex,
	}
}

func MapToCategoryListResponse(categories []domain.Category) []CategoryResponse {
	res := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		res[i] = MapToCategoryResponse(&cat)
	}
	return res
}
