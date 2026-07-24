package handler

import (
	"io"
	"math"
	"strconv"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/Jaruvat303/cashlog/pkg/response"
	"github.com/Jaruvat303/cashlog/pkg/timeutil"
	"github.com/Jaruvat303/cashlog/pkg/validate"
	"github.com/gofiber/fiber/v2"
)

type TransactionHandler struct {
	txUsecase domain.TransactionUsecase
	log       logger.Logger
}

// NewTransactionHandler สร้าง TransactionHandler ใหม่
func NewTransactionHandler(txUsecase domain.TransactionUsecase, applogger logger.Logger) *TransactionHandler {
	return &TransactionHandler{
		txUsecase: txUsecase,
		log:       applogger,
	}
}

// UplaodSlipAndLog สำหรับอัปโหลดสลิปและบันทึก Transaction
// @Summary อัปโหลดสลิปและบันทึก Transaction
// @Description อัปโหลดสลิปและบันทึก Transaction โดยต้องแนบไฟล์สลิปและระบุชื่อไฟล์ต้นฉบับ
// @Tags Transaction
// @Accept multipart/form-data
// @Produce json
// @Param local_image_name formData string true "ชื่อไฟล์ต้นฉบับของสลิป"
// @Param image formData file true "ไฟล์สลิป (รูปภาพ)"
// @Success 201 {object} response.JsonResponse[dto.CategoryResponse] "Transaction processed successfully"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request  <br>error_code: INVALID_INPUT_PARAMETERS <br>message: 1. Missing required form field 'local_image_name' 2. Missing required file 'image' in multipart/form-data request."
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /transactions/upload-slip [post]
func (h *TransactionHandler) UplaodSlipAndLog(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// อ่านค่าชื่อไฟล์ภาพต้นฉบับจาก Form Value
	localImageName := c.FormValue("local_image_name")
	if localImageName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required form field 'local_image_name'.")
	}

	// ดึงไฟล์ภาพสลืปที่แนบมาในชื่อฟิลด์ "image"
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required file 'image' in multipart/form-data request.")
	}

	// เปิดไฟล์และแปลงข้อมูลภาพให้อยู่่ในรูปแบบ byte array ([]byte)
	file, err := fileHeader.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Unable to read the uploaded image. The file may be corrupted or incomplete.")
	}
	// ปิดไฟล์เมื่อฟังก์ชันนี้ทำงานเสร็จ
	defer func() {
		_ = file.Close()
	}()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Unable to read the uploaded image. The file may be corrupted or incomplete.")
	}

	// ส่งข้อมูลรูปภาพและชื่อไฟล์เข้าไปให้ Usecase ประมวลผลตรรกะทั้งหมด
	resultTx, err := h.txUsecase.SyncTransaction(ctx, imageBytes, localImageName)
	if err != nil {
		return err
	}

	// เคสที่งานสำเร็จแบบพิเศษข้ามเพราะข้อมูลซ้ำ
	if resultTx == nil {
		return response.OkMessage(c, fiber.StatusOK,
			"Transaction processed successfully (skipped or duplicate caught early)")
	}

	// map ข้อมูลจาก Domain Model เป็น DTO สำหรับส่งกลับไปให้ Client
	dtoResult := dto.MapToTransactionResponse(resultTx)

	return response.Success(c, fiber.StatusCreated, "Transaction processed successfully", dtoResult)
}

// GetDashboardSummary ตอบกลับข้อมูลสรุปรายรับ-รายจ่ายประจำเดือนหรือรายปี
// @Summary ดึงข้อมูลสรุปรายรับ-รายจ่ายประจำเดือนหรือรายปี
// @Description ดึงข้อมูลสรุปรายรับ-รายจ่ายประจำเดือนหรือรายปี โดยสามารถระบุช่วงเวลาได้ผ่าน Query Parameters
// @Tags Transaction
// @Accept json
// @Produce json
// @Param scope query string false "ช่วงเวลาที่ต้องการดึงข้อมูล" Enums(monthly, yearly) default(monthly)
// @Param month query int false "เดือนที่ต้องการดึงข้อมูล (1-12)" default(current month)
// @Param year query int false "ปีที่ต้องการดึงข้อมูล" default(current year)
// @Success 200 {object} response.JsonResponse[dto.DashboardSummaryResponse] "Dashboard summary fetched successfully"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request  <br>error_code: INVALID_INPUT_PARAMETERS <br>message: Invalid query parameter 'scope'. Allowed values are 'monthly' or 'yearly'."
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /transactions/summary [get]
func (h *TransactionHandler) GetDashboardSummary(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// ดึง Query Parameters พร้อมกำหนดค่า Default
	now := timeutil.NowInBangkok()
	scope := c.Query("scope", "monthly") // กำหนดค่าพื้นฐานรายเดือน
	month := c.QueryInt("month", int(now.Month()))
	year := c.QueryInt("year", now.Year())

	// ตรวจสอบข้อความที่ client ส่งมา
	if scope != "monthly" && scope != "yearly" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid query parameter 'scope'. Allowed values are 'monthly' or 'yearly'.")
	}

	// เรียกใช้งาน Usecase เพื่อดึงข้อมูล
	summary, err := h.txUsecase.GetDashboardSummary(ctx, scope, month, year)
	if err != nil {
		return err
	}

	// map ข้อมูลจาก Domain Model เป็น DTO สำหรับส่งกลับไปให้ Client
	dtoSummary := dto.MapToDashboardSummaryResponse(summary)

	return response.Success(c, fiber.StatusOK, "Dashboard summary fetched successfully", dtoSummary)
}

// GetMonthlyHistory สำหรับดึงข้อมูล Transaction ตามเวลาที่กำหนด
// @Summary ดึงข้อมูล Transaction ตามเวลาที่กำหนด
// @Description ดึงข้อมูล Transaction ตามเวลาที่กำหนด โดยสามารถระบุปี เดือน และการแบ่งหน้า (Pagination) ได้ผ่าน Query Parameters
// @Tags Transaction
// @Accept json
// @Produce json
// @Param year query int false "ปีที่ต้องการดึงข้อมูล" default(current year)
// @Param month query int false "เดือนที่ต้องการดึงข้อมูล (1-12)" default(current month)
// @Param page query int false "หน้าที่ต้องการดึงข้อมูล" default(1)
// @Param limit query int false "จำนวนรายการต่อหน้า" default(20)
// @Success 200 {object} response.PaginatedResponse[dto.TransactionResponse] "ดึงข้อมูล Transaction สำเร็จ"
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /transactions/ [get]
func (h *TransactionHandler) GetMonthlyHistory(c *fiber.Ctx) error {
	// 1. แกะค่าจาก Query Parameters (พร้อมกำหนดค่า Default เผื่อหน้าบ้านไม่ได้ส่งมา)
	year, _ := strconv.Atoi(c.Query("year", "0"))
	month, _ := strconv.Atoi(c.Query("month", "0"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	ctx := c.UserContext()
	transactions, err := h.txUsecase.FetchTransactions(ctx, domain.FetchTransactionInput{
		Month: month,
		Year:  year,
		Limit: limit,
		Page:  page,
	})
	if err != nil {
		return err
	}

	// แปลงข้อมูลจาก Domain Model เป็น DTO สำหรับส่งกลับไปให้ Client
	dtos := dto.MapToTransactionListResponse(transactions.Transactions)

	// ตำนนจำนวนหน้าที่มีทั้งหมด
	totalPages := int(math.Ceil(float64(transactions.TotalItems) / float64(limit)))
	if totalPages == 0 && transactions.TotalItems == 0 {
		totalPages = 1
	}

	// ใส่ข้อมูล Meta สำหรับ Pagination
	meta := response.PaginationMeta{
		TotalItems:  int(transactions.TotalItems),
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    limit,
	}

	return response.Paginated(c, fiber.StatusOK, "Monthly history fetched successfully", dtos, meta)
}

// UpdateTransaction สำหรับแก้ไขข้อมูล Transaction
// @Summary แก้ไขข้อมูล Transaction
// @Description แก้ไขข้อมูล Transaction โดยต้องระบุ ID ของ Transaction ที่ต้องการแก้ไข และสามารถแก้ไขข้อมูลบางส่วนได้ (Partial Update)
// @Tags Transaction
// @Accept json
// @Produce json
// @Param id path int true "ID ของ Transaction ที่ต้องการแก้ไข"
// @Param request body dto.UpdateTransactionInput true "ข้อมูลที่ต้องการแก้ไข (Partial Update)"
// @Success 200 {object} response.JsonResponse[dto.TransactionResponse] "Transaction updated successfully"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request  <br>error_code: INVALID_INPUT_PARAMETERS <br>message: 1. Invalid ID format. The path parameter 'id' must be a positive integer. 2. Invalid JSON format. 3. Validation failed for the request data"
// @Failure 499 {object} dto.ErrorResponseDTO "Client Closed Request  <br>error_code: REQUEST_CANCELED <br>message: The request was canceled by the user"
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /transactions/{id} [patch]
func (h *TransactionHandler) UpdateTransaction(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// แกะ ID จาก URL Path แปลงเป็น uint
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ID format. The path parameter 'id' must be a positive integer.")
	}

	// ผูกข้อมูล (Bind) JSON Request Body เข้ากับ DTO
	var input dto.UpdateTransactionInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format.")
	}

	// ตรวจสอบ Tag Validate
	if err := validate.ValidateStruct(input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed for the request data.")
	}

	result, err := h.txUsecase.UpdateTransaction(ctx, uint(id), input.ToDomainUpdateParam())
	if err != nil {
		return err
	}

	// Map ข้อมูลจาก Domain Model เป็น DTO สำหรับส่งกลับไปให้ Client
	resultDTO := dto.MapToTransactionResponse(result)

	return response.Success(c, fiber.StatusOK, "Transaction updated successfully", resultDTO)
}

// Delete Transaction สำหรับการลบข้อมูล ด้วย id (soft delete)
// @Summary ลบข้อมูล Transaction
// @Description ลบข้อมูล Transaction โดยต้องระบุ ID ของ Transaction ที่ต้องการลบ
// @Tags Transaction
// @Accept json
// @Produce json
// @Param id path int true "ID ของ Transaction ที่ต้องการลบ"
// @Success 200 {object} response.JsonResponse[dto.EmptyData] "Transaction deleted successfully"
// @Failure 400 {object} dto.ErrorResponseDTO "Bad Request  <br>error_code: INVALID_INPUT_PARAMETERS <br>message: 1. Invalid ID format. The path parameter 'id' must be a positive integer."
// @Failure 500 {object} dto.ErrorResponseDTO "Internal Server Error <br>error_code: INTERNAL_SERVER_ERROR or INTERNAL_DATABASE_ERROR <br>message: Something went wrong, please try again later"
// @Failure 504 {object} dto.ErrorResponseDTO "Gateway Timeout <br>error_code: DATABASE_TIMEOUT <br>message: The database operation timed out, please try again"
// @Router /transactions/{id} [delete]
func (h *TransactionHandler) DeleteTransaction(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// แกะ ID จาก URL Path แปลงเป็น uint
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ID format. The path parameter 'id' must be a positive integer.")
	}

	// เรียกใช้งาน transaction usecase
	err = h.txUsecase.DeleteTransaction(ctx, uint(id))
	if err != nil {
		return err
	}

	return response.OkMessage(c, fiber.StatusOK, "Transaction deleted successfully")

}
