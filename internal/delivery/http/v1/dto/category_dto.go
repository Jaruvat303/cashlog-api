package dto

// CreateCategoryInput
type CreateCategoryInput struct {
	Name    string  `json:"name" validate:"required,min=2,max=50"`
	Type    string  `json:"type" validate:"required,oneof=INCOME EXPENSE"`
}

// UpdateCategoryInput
type UpdateCategoryInput struct {
	Name    *string `json:"name" validate:"omitempty,min=3,max=50"`
}
