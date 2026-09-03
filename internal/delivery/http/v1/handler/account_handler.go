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

type AccountHandler struct {
	usecase domain.AccountUsecase
	log     logger.Logger
}

func NewAccountHandler(usecase domain.AccountUsecase, appLogger logger.Logger) *AccountHandler {
	return &AccountHandler{
		usecase: usecase,
		log:     appLogger,
	}
}

// CreateAccount godoc
// @Summary สร้าง Account ใหม่
// @Description สร้างบัญชีใหม่ โดยต้องระบุชื่อและประเภทของบัญชี
// @Tags Account
// @Accept json
// @Produce json
// @Param input body dto.CreateAccountInput true "ข้อมูลสำหรับสร้าง Account"
// @Success 201 {object} response.JsonResponse[dto.AccountResponse] "create account successfull"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request  <br>error_code: INVALID_INPUT_PARAMETERS <br>message: 1. Invalid JSON format.  2. Validation failed for the request data."
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /accounts [post]
func (h *AccountHandler) CreateAccount(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// ผูก Json Body กับ DTO data
	var input dto.CreateAccountInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format.")
	}

	// ตรวจสอบ validate tag
	if err := validate.ValidateStruct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed for the request data.")
	}

	// เรียกงาน usecase
	account, err := h.usecase.CreateAccount(ctx, input.ToDomainCreateParam())
	if err != nil {
		return err
	}

	// แปลงข้อมูลจาก Domain เป็น DTO Response
	dtoResponse := dto.MapToAccountResponse(account)

	return response.Success(c, fiber.StatusCreated, "create account successfull", dtoResponse)
}

// UpdateAccount godoc
// @Summary อัปเดต Account
// @Description อัปเดตข้อมูลบัญชี โดยต้องระบุ ID ของบัญชีที่ต้องการอัปเดต
// @Tags Account
// @Accept json
// @Produce json
// @Param id path int true "ID ของ Account ที่ต้องการอัปเดต"
// @Param input body dto.UpdateAccountInput true "ข้อมูลสำหรับอัปเดต Account"
// @Success 200 {object} response.JsonResponse[dto.AccountResponse] "update account successfull"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request  <br>error_code: INVALID_INPUT_PARAMETERS <br>message: 1. Invalid ID format (must be a positive integer) 2. Invalid JSON format 3. Validation failed for the request data"
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /accounts/{id} [patch]
func (h *AccountHandler) UpdateAccount(c *fiber.Ctx) error {
	ctx := c.UserContext()

	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ID format. The path parameter 'id' must be a positive integer.")
	}

	// ผูก Json Body กับ DTO data
	var input dto.UpdateAccountInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format.")
	}

	// ตรวจสอบ Validate tag
	if err := validate.ValidateStruct(input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed for the request data.")
	}

	// เรียกใช้งาน usecase
	account, err := h.usecase.UpdateAccount(ctx, uint(id), input.ToDomainUpdateParam())
	if err != nil {
		return err
	}

	// แปลงข้อมูลจาก Domain เป็น DTO Response
	dtoResponse := dto.MapToAccountResponse(account)

	return response.Success(c, fiber.StatusOK, "update account successfull", dtoResponse)
}

// DeleteAccount godoc
// @Summary ลบ Account
// @Description ปิดใช้งานบัญชี (soft delete) โดยต้องระบุ ID ของบัญชีที่ต้องการลบ
// @Tags Account
// @Accept json
// @Produce json
// @Param id path int true "ID ของ Account ที่ต้องการลบ"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request <br>error_code: INVALID_INPUT_PARAMETERS <br>Message: Invalid ID format. The path parameter 'id' must be a positive integer."
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /accounts/{id} [delete]
func (h *AccountHandler) DeleteAccount(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// รับ id parameter แล้วแปลงเป็น unit
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ID format. The path parameter 'id' must be a positive integer.")
	}

	// เรียกใช้งาน usecase
	if err := h.usecase.DeleteAccount(ctx, uint(id)); err != nil {
		return err
	}

	return response.OkMessage(c, fiber.StatusOK, "delete account successfull")
}

// FetchActiveAccounts godoc
// @Summary ดึงรายการ Account ที่ใช้งานอยู่
// @Description ดึงรายการบัญชีทั้งหมดที่ is_active=true
// @Tags Account
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonResponse[[]dto.AccountResponse] "FetchActiveAccounts successful <br>message: FetchActiveAccounts successfull"
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /accounts [get]
func (h *AccountHandler) FetchActiveAccounts(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// เรียกใช้งาน Usecase
	accounts, err := h.usecase.FetchActiveAccounts(ctx)
	if err != nil {
		return err
	}

	// แปลงข้อมูลจาก Domain เป็น DTO Response
	dtoAccounts := dto.MapToAccountListResponse(accounts)

	return response.Success(c, fiber.StatusOK, "FetchActiveAccounts successfull", dtoAccounts)
}
