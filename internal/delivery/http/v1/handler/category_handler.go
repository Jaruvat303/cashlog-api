package handler

import (
	"strconv"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
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

func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// ผูก Json Body กับ DTO data
	var input dto.CreateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// ตรวจสอบ validate tag
	if err := validate.ValidateStruct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "validate struct error body")
	}

	// เรียกงาน usecase
	if err := h.usecase.CreateCategory(ctx, input); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "create category successfull",
	})
}

func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()

	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "parse id param to uint error")
	}

	// ผูก Json Body กับ DTO data
	var input dto.UpdateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// ตรวจสอบ Validate tag
	if err := validate.ValidateStruct(input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input struct")
	}

	// เรียกใช้งาน usecase
	category, err := h.usecase.UpdateCategory(ctx, uint(id), input)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "update category successfull",
		"data":    category,
	})
}

func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// รับ id parameter แล้วแปลงเป็น unit
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "parse id param to uint error")
	}

	// เรียกใช้งาน usecase
	if err := h.usecase.DeleteCategory(ctx, uint(id)); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "delete category successfull",
	})

}

func (h *CategoryHandler) FetchCategories(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// เรียกใช้งาน Usecase
	categories, err := h.usecase.FetchCategories(ctx)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    categories,
	})
}
