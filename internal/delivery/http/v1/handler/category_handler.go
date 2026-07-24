package handler

import (
	"strconv"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/Jaruvat303/cashlog/pkg/response"
	"github.com/Jaruvat303/cashlog/pkg/validate"
	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
	usecase domain.CategoryUsecase
	log     logger.Logger
}

func NewCategoryHandler(usecase domain.CategoryUsecase, appLogger logger.Logger) *CategoryHandler {
	return &CategoryHandler{
		usecase: usecase,
		log:     appLogger,
	}
}

// CreateCategory godoc
// @Summary สร้าง Category ใหม่
// @Description สร้าง Category ใหม่ โดยต้องระบุชื่อและประเภทของ Category
// @Tags Category
// @Accept json
// @Produce json
// @Param input body dto.CreateCategoryInput true "ข้อมูลสำหรับสร้าง Category"
// @Success 201 {object} response.JsonResponse[dto.CategoryResponse] "create category successfull"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request  <br>error_code: INVALID_INPUT_PARAMETERS <br>message: 1. Invalid JSON format.  2. Validation failed for the request data."
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// ผูก Json Body กับ DTO data
	var input dto.CreateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format.")
	}

	// ตรวจสอบ validate tag
	if err := validate.ValidateStruct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed for the request data.")
	}

	// เรียกงาน usecase
	category, err := h.usecase.CreateCategory(ctx, input.ToDomainCreateParam())
	if err != nil {
		return err
	}

	// แปลงข้อมูลจาก Domain เป็น DTO Response
	dtoResponse := dto.MapToCategoryResponse(category)

	return response.Success(c, fiber.StatusCreated, "create category successfull", dtoResponse)
}

// UpdateCategory godoc
// @Summary อัปเดต Category
// @Description อัปเดตข้อมูล Category โดยต้องระบุ ID ของ Category ที่ต้องการอัปเดต
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "ID ของ Category ที่ต้องการอัปเดต"
// @Param input body dto.UpdateCategoryInput true "ข้อมูลสำหรับอัปเดต Category"
// @Success 200 {object} response.JsonResponse[dto.CategoryResponse] "update category successfull"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request  <br>error_code: INVALID_INPUT_PARAMETERS <br>message: 1. Invalid ID format (must be a positive integer) 2. Invalid JSON format 3. Validation failed for the request data"
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /categories/{id} [patch]
func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()

	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ID format. The path parameter 'id' must be a positive integer.")
	}

	// ผูก Json Body กับ DTO data
	var input dto.UpdateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format.")
	}

	// ตรวจสอบ Validate tag
	if err := validate.ValidateStruct(input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed for the request data.")
	}

	// เรียกใช้งาน usecase
	category, err := h.usecase.UpdateCategory(ctx, uint(id), input.ToDomainUpdateParam())
	if err != nil {
		return err
	}

	// แปลงข้อมูลจาก Domain เป็น DTO Response
	dtoResponse := dto.MapToCategoryResponse(category)

	return response.Success(c, fiber.StatusOK, "update category successfull", dtoResponse)
}

// DeleteCategory godoc
// @Summary ลบ Category
// @Description ลบ Category โดยต้องระบุ ID ของ Category ที่ต้องการลบ
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "ID ของ Category ที่ต้องการลบ"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request <br>error_code: INVALID_INPUT_PARAMETERS <br>Message: Invalid ID format. The path parameter 'id' must be a positive integer."
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// รับ id parameter แล้วแปลงเป็น unit
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ID format. The path parameter 'id' must be a positive integer.")
	}

	// เรียกใช้งาน usecase
	if err := h.usecase.DeleteCategory(ctx, uint(id)); err != nil {
		return err
	}

	return response.OkMessage(c, fiber.StatusOK, "delete category successfull")

}

// FetchCategoriesByType godoc
// @Summary ดึงข้อมูล Category ตามประเภท
// @Description ดึงข้อมูล Category ตามประเภท โดยต้องระบุประเภทของ Category (income หรือ expense)
// @Tags Category
// @Accept json
// @Produce json
// @Param type query string false "ประเภทของ Category" Enums(income, expense) default(expense)
// @Success 200 {object} response.JsonResponse[[]dto.CategoryResponse] "FetchCategoriesByType successful <br>message: FetchCategoriesByType successfull"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request <br>error_code: INVALID_INPUT_PARAMETERS <br>message: Invalid query parameter 'type'. Allowed values are 'expense' or 'income'."
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /categories [get]
func (h *CategoryHandler) FetchCategoriesByType(c *fiber.Ctx) error {
	ctx := c.UserContext()

	types := c.Query("type", "expense")
	if types != "expense" && types != "income" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid query parameter 'type'. Allowed values are 'expense' or 'income'.")
	}

	// เรียกใช้งาน Usecase
	categories, err := h.usecase.FetchCategoriesByType(ctx, types)
	if err != nil {
		return err
	}

	// แปลงข้อมูลจาก Domain เป็น DTO Response
	dtoCategories := dto.MapToCategoryListResponse(categories)

	return response.Success(c, fiber.StatusOK, "FetchCategoriesByType successfull", dtoCategories)
}
