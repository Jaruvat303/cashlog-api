package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/handler"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDashboardSummary(t *testing.T) {
	// ข้อมูลจำลองสำหรับตรวจเช็ต Response Body
	mockSummary := &domain.DashboardSummary{
		TotalIncome:  15000.0,
		TotalExpense: 5000.0,
		Scope:        "monthly",
		Year:         2026,
		Month:        6,
	}

	tests := []struct {
		name           string
		url            string
		setupMock      func(uc *domain.TransactionUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "1. Validation Failed - Scope Parameter ไม่ถูกต้อง",
			url:  "/api/v1/dashboard?scope=daily",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// usecase ต้องไม่ทำงาน ในก่ารทดสอบนี
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "invalid scope parameter",
		},
		{
			name: "2. Success - รับค่าพารามิเตอร์ปกติ และ usecase ทำงานสำเร็จ",
			url:  "/api/v1/dashboard?scope=monthly&month=6&year=2026",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// จำลองการทำงานของ Usecase
				uc.On("GetDashboardSummary", mock.Anything, "monthly", 6, 2026).
					Return(mockSummary, nil)
			},
			expectedStatus: fiber.StatusOK,
			// เนื่องจากเรา unittest ในส่วนของ usecase แล้วข้อมูลใน field data เลยไม่จำเป็นต้อง test อีก
			expectedBody: `"success":true`,
		},
		{
			name: "3. Usecase Error - หลังบ้านพัง Handler ส่งต่อ Error ได้",
			url:  "/api/v1/dashboard?scope=monthly&month=6&year=2026",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// จำลองการทำงานของ Usecase
				uc.On("GetDashboardSummary", mock.Anything, "monthly", 6, 2026).
					Return(nil, errors.New("something went wrong in usecase"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "something went wrong in usecase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			app := fiber.New()
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()

			tt.setupMock(mockUsecase)

			// ผูก Route ้เข้ากับ Handler function
			txHandler := handler.NewTransactionHandler(mockUsecase, mockLog)
			app.Get("/api/v1/dashboard", txHandler.GetDashboardSummary)

			// จำลอง HTTP Request เสทือนส่งมาจาก Client
			req := httptest.NewRequest("GET", tt.url, nil)
			req.Header.Set("Content-Type", "application/json")

			// Act ส่ง Request เข้าไปในระบบ fiber ผ่านตำสั่ง app.Test()
			resp, err := app.Test(req, -1) // -1 คือปิดการทำงานของ timeout

			// Assert ตรวจสอบ HTTP StatusCode
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Assert ข้่อมูล JSON Response Body
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyStr := string(bodyBytes)
			assert.Contains(t, bodyStr, tt.expectedBody)

			// ตรวจสอบความถูกต้องว่า Usecase ถูกทำงานตามแผนจริงไหม
			mockUsecase.AssertExpectations(t)
		})
	}

}

func TestGetMonthlyHistory(t *testing.T) {
	mockHistory := []domain.Transaction{
		{ID: 1, Amount: 200},
		{ID: 2, Amount: 500},
	}

	tests := []struct {
		name           string
		url            string
		setupMock      func(uc *domain.TransactionUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "1. Success - รับค่าพารามิเตอร์ปกติ เรียกดูข้อมูลสำเร็จ",
			url:  "/api/v1/transaction?month=6&year=2026",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("GetMonthlyHistory", mock.Anything, 6, 2026).Return(mockHistory, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name: "2. Usecase Error - หลังบ้านพัง Handler ส่งต่อ Error ได้",
			url:  "/api/v1/transaction?month=6&year=2026",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("GetMonthlyHistory", mock.Anything, 6, 2026).Return(nil, errors.New("something went wrong in usecase"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "something went wrong in usecase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			app := fiber.New()
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()

			tt.setupMock(mockUsecase)

			// ผูก Rote เข้ากับ Handler function
			txHandler := handler.NewTransactionHandler(mockUsecase, mockLog)
			app.Get("/api/v1/transaction", txHandler.GetMonthlyHistory)

			// สร้าง Request ยืงไปที่ Route
			req := httptest.NewRequest("GET", tt.url, nil)
			req.Header.Set("Content-Type", "application/json")

			// Act ส่ง request เข้าใปใน fiber app
			resp, err := app.Test(req, -1)

			// Assert ตรวจสอบ Response Code
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Assert ตรวจสอบ ResponseBody
			bodyByte, err := io.ReadAll(resp.Body)
			bodyStr := string(bodyByte)
			assert.Contains(t, bodyStr, tt.expectedBody)

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestUpdateTransaction(t *testing.T) {
	fixedTime := time.Date(2026, time.June, 19, 16, 0, 0, 0, time.Local)

	// ข้อมูลเข้า usecase
	validInput := dto.UpdateTransactionInput{
		Amount:          pkg.PTR(200.0),
		Note:            pkg.PTR("edit amount"),
		CategoryID:      pkg.PTR(int64(3)),
		TransactionDate: pkg.PTR(fixedTime),
	}

	// เตรียมข้อมูลผลลัพธ์ที่จะได้จาก Usecase
	mockResult := &domain.Transaction{
		ID:              1,
		Amount:          200,
		CategoryID:      3,
		Note:            "edit amount",
		TransactionDate: fixedTime,
	}

	tests := []struct {
		name           string
		paramID        string
		requestBody    interface{}
		setupMock      func(uc *domain.TransactionUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "1. Bad Request - ID ใน URL ไม่ใช่ตัวเลข",
			paramID:     "one",
			requestBody: validInput,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// usecase จะยังไม่ทำงาน
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "parse id param to uint error",
		},
		{
			name:        "2. Bad Request - JSON Body รูปแบบไม่ถูกต้อง แกะข้อมูลไม่ได้",
			paramID:     "1",
			requestBody: "{invalid json}",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// usecase จะยังไม่ทำงาน
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:    "3. Bad Request - ข้อมูลไม่ผ่าน Tag Validation",
			paramID: "1",
			requestBody: dto.UpdateTransactionInput{
				Amount: pkg.PTR(-50.0), // สมมติว่าใน DTO ห้ามติดลบ ทำให้ ValidateStruct พัง
			},
			setupMock:      func(uc *domain.TransactionUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "validation struct error",
		},

		{
			name:        "4. Internal Server Error - เกิดข้อผิดพลาดที่ Usecase Layer",
			paramID:     "1",
			requestBody: validInput,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("UpdateTransaction", mock.Anything, uint(1), validInput).Return(nil, domain.ErrInternalDB)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "internal database error",
		},
		{
			name:        "5. Success - อัปเดตข้อมูลสำเร็จ",
			paramID:     "1",
			requestBody: validInput,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("UpdateTransaction", mock.Anything, uint(1), validInput).Return(mockResult, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			app := fiber.New()
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()
			tt.setupMock(mockUsecase)

			// ผูก route เข้ากับ Handler
			handler := handler.NewTransactionHandler(mockUsecase, mockLog)
			app.Patch("/api/v1/transactions/:id", handler.UpdateTransaction)

			// แปลง Request Body ให้อยู่ในรูปของ Bytes.Buffer
			var jsonBody []byte
			if strBody, ok := tt.requestBody.(string); ok {
				jsonBody = []byte(strBody)
			} else {
				jsonBody, _ = json.Marshal(tt.requestBody)
			}

			// สร้าง request ยิงไปที่ route
			req := httptest.NewRequest(fiber.MethodPatch, "/api/v1/transactions/"+tt.paramID, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Act
			resp, err := app.Test(req, -1)

			// Assert Status code
			assert.NoError(t, err)
			assert.Equal(t, resp.StatusCode, tt.expectedStatus)

			// Assert Response Body
			bodyByte, err := io.ReadAll(resp.Body)
			bodyStr := string(bodyByte)
			assert.Contains(t, bodyStr, tt.expectedBody)

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestDeleteTransaction(t *testing.T) {
	tests := []struct {
		name           string
		paramID        string
		setupMock      func(uc *domain.TransactionUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "1. Bad Request - ID ใน URL ไม่ใช่ตัวเลข",
			paramID: "one",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// usecase จะยังไม่ทำงาน
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "parse id param to uint error",
		},

		{
			name:    "2. Internal Server Error - เกิดข้อผิดพลาดที่ Usecase Layer",
			paramID: "1",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("DeleteTransaction", mock.Anything, uint(1)).Return(domain.ErrInternalDB)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "internal database error",
		},
		{
			name:    "3. Success - ลบข้อมูลสำเร็จ",
			paramID: "1",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("DeleteTransaction", mock.Anything, uint(1)).Return(nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			app := fiber.New()
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()
			tt.setupMock(mockUsecase)

			// ผูก route เข้ากับ Handler
			handler := handler.NewTransactionHandler(mockUsecase, mockLog)
			app.Delete("/api/v1/transactions/:id", handler.DeleteTransaction)

			// สร้าง request ยิงไปที่ route
			req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/transactions/"+tt.paramID, nil)
			req.Header.Set("Content-Type", "application/json")

			// Act
			resp, err := app.Test(req, -1)

			// Assert Status code
			assert.NoError(t, err)
			assert.Equal(t, resp.StatusCode, tt.expectedStatus)

			// Assert Response Body
			bodyByte, err := io.ReadAll(resp.Body)
			bodyStr := string(bodyByte)
			assert.Contains(t, bodyStr, tt.expectedBody)

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestUplaodSlipAndLog(t *testing.T) {
	url := "/api/v1/log"
	tx := &domain.Transaction{
		ID:     1,
		Amount: 200,
	}

	tests := []struct {
		name           string
		url            string
		imageName      string
		hasImage       bool
		setupMock      func(uc *domain.TransactionUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "1. Validation Failed - image name field ไม่มีข้อมูล",
			imageName: "",
			hasImage:  true,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// usecase จะยังไม่ทำงานใน case นี้
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "The field `local_image_name` is required.",
		},
		{
			name:      "2. Validation Failed - on image file",
			imageName: "slip_06.jpg",
			hasImage:  false,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// usecase จะไม่ทำงานใน case นี้
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "The `image` flie is required in multipart/form-data.",
		},
		{
			name:      "3. Succes - Create Sucess New Data",
			imageName: "slip_06.jpg",
			hasImage:  true,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				mockImageByte := []byte("fake-image-bytes")
				uc.On("SyncTransaction", mock.Anything, mockImageByte, "slip_06.jpg").Return(tx, nil)
			},
			expectedStatus: fiber.StatusCreated,
			expectedBody:   `"message":"Transaction logged successfully"`,
		},
		{
			name:      "4. Succes - Create Sucess Duplicate Data",
			imageName: "slip_06.jpg",
			hasImage:  true,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				mockImageByte := []byte("fake-image-bytes")
				uc.On("SyncTransaction", mock.Anything, mockImageByte, "slip_06.jpg").Return(nil, nil)
			},
			expectedStatus: fiber.StatusCreated,
			expectedBody:   `"message":"Transaction processed successfully (skipped or duplicate caught early)"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			app := fiber.New()
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()

			tt.setupMock(mockUsecase)

			txHandler := handler.NewTransactionHandler(mockUsecase, mockLog)
			app.Post(url, txHandler.UplaodSlipAndLog)

			// จำลอง multipart/formdata ใน Unit Test
			body := &bytes.Buffer{}

			writer := multipart.NewWriter(body)
			// ใส่ Field Text
			_ = writer.WriteField("local_image_name", tt.imageName)
			if tt.hasImage {
				// ใส่ Field รูปภาพ
				part, _ := writer.CreateFormFile("image", tt.imageName)
				_, _ = part.Write([]byte("fake-image-bytes"))
			}

			_ = writer.Close()

			// สร้าง Request โดยเอา Body และใส่ 'Content-Type จาก writer
			req := httptest.NewRequest("POST", url, body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Act ส่ง Request เข้าไปใน Fiber App
			resp, err := app.Test(req, -1)

			// Assert ตรวจสอบ Http Status
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Assert ตรวจสอบ Response Body
			bodyByte, _ := io.ReadAll(resp.Body)
			bodyStr := string(bodyByte)
			assert.Contains(t, bodyStr, tt.expectedBody)

			mockUsecase.AssertExpectations(t)

		})
	}
}
