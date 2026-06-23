package handler

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/Jaruvat303/cashlog/pkg/timeutil"
	"github.com/Jaruvat303/cashlog/pkg/validate"
	"github.com/gofiber/fiber/v2"
)

type TransactionHandler struct {
	txUsecase domain.TransactionUsecase
	log       logger.Logger
}

// NewTransactionHandler ทำหน้าที่สร้าง Handler เพื่อนำไปผูกกับ Router ของ Fiber
func NewTransactionHandler(txUsecase domain.TransactionUsecase, applogger logger.Logger) *TransactionHandler {
	return &TransactionHandler{
		txUsecase: txUsecase,
		log:       applogger,
	}
}

// UploadSlipAndLog ทำหน้าที่ไฟล์ภาพสลืปและข้อมูลเพื่อนำไปประมวลผลบันทึกรายรับรายจ่าย
func (h *TransactionHandler) UplaodSlipAndLog(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// อ่านค่าชื่อไฟล์ภาพต้นฉบับจาก Form Value
	localImageName := c.FormValue("local_image_name")
	if localImageName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "The field `local_image_name` is required.")
	}

	// ดึงไฟล์ภาพสลืปที่แนบมาในชื่อฟิลด์ "image"
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "The `image` flie is required in multipart/form-data.")
	}

	// เปิดไฟล์และแปลงข้อมูลภาพให้อยู่่ในรูปแบบ byte array ([]byte)
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open file image: %w", err)
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read image byte: %w", err)
	}

	// ส่งข้อมูลรูปภาพและชื่อไฟล์เข้าไปให้ Usecase ประมวลผลตรรกะทั้งหมด
	resultTx, err := h.txUsecase.SyncTransaction(ctx, imageBytes, localImageName)
	if err != nil {
		return err
	}

	// เคสที่งานสำเร็จแบบพิเศษข้ามเพราะข้อมูลซ้ำ
	if resultTx == nil {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Transaction processed successfully (skipped or duplicate caught early)",
			"data":    resultTx,
		})
	}

	// เคสงานสำเร็จแบบปกติ (Happy Path แบบมีข้อมูลใหม่)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Transaction logged successfully",
		"data":    resultTx,
	})
}

// GetDashboardSummary ตอบกลับข้อมูลสรุปรายรับ-รายจ่ายประจำเดือนหรือรายปี
func (h *TransactionHandler) GetDashboardSummary(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// ดึง Query Parameters พร้อมกำหนดค่า Default
	now := timeutil.NowInBangkok()
	scope := c.Query("scope", "monthly") // กำหนดค่าพื้นฐานรายเดือน
	month := c.QueryInt("month", int(now.Month()))
	year := c.QueryInt("year", now.Year())

	// ตรวจสอบข้อความที่ client ส่งมา
	if scope != "monthly" && scope != "yearly" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid scope parameter")
	}

	// เรียกใช้งาน Usecase เพื่อดึงข้อมูล
	summary, err := h.txUsecase.GetDashboardSummary(ctx, scope, month, year)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    summary,
	})
}

// GetMonthlyHistory สำหรับดึงข้อมูล Transaction ตามเวลาที่กำหนด
func (h *TransactionHandler) GetMonthlyHistory(c *fiber.Ctx) error {
	now := timeutil.NowInBangkok()
	month := c.QueryInt("month", int(now.Month()))
	year := c.QueryInt("year", now.Year())

	ctx := c.UserContext()
	transactions, err := h.txUsecase.GetMonthlyHistory(ctx, month, year)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    transactions,
	})
}

// UpdateTransaction สำหรับแก้ไขข้อมูล Transaction
func (h *TransactionHandler) UpdateTransaction(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// แกะ ID จาก URL Path แปลงเป็น uint
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "parse id param to uint error")
	}

	// ผูกข้อมูล (Bind) JSON Request Body เข้ากับ DTO
	var input dto.UpdateTransactionInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// ตรวจสอบ Tag Validate
	if err := validate.ValidateStruct(input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "validation struct error")
	}

	result, err := h.txUsecase.UpdateTransaction(ctx, uint(id), input)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "update transaction successfully",
		"data":    result,
	})
}

// Delete Transaction สำหรับการลบข้อมูล ด้วย id (soft delete)
func (h *TransactionHandler) DeleteTransaction(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// แกะ ID จาก URL Path แปลงเป็น uint
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "parse id param to uint error")
	}

	// เรียกใช้งาน transaction usecase
	err = h.txUsecase.DeleteTransaction(ctx, uint(id))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":        true,
		"message":        "delete transaction successfull",
		"transaction_id": id,
	})

}
