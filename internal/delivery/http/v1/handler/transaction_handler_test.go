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
	"github.com/Jaruvat303/cashlog/pkg/timeutil"
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
			expectedBody:   "Invalid query parameter 'scope'. Allowed values are 'monthly' or 'yearly'.",
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
			expectedBody:   "Something went wrong, please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)

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

	mockInput := domain.FetchTransactionInput{
		Month: 6,
		Year:  2026,
		Page:  1,
		Limit: 10,
	}
	mockHistory := []domain.Transaction{
		{ID: 1, Amount: 200},
		{ID: 2, Amount: 500},
	}

	mockUsecaseResponse := &domain.FetchTransactionResult{
		Transactions: mockHistory,
		TotalItems:   int64(len(mockHistory)),
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
			url:  "/api/v1/transaction?month=6&year=2026&page=1&limit=10",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("FetchTransactions", mock.Anything, mockInput).Return(mockUsecaseResponse, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name: "2. Usecase Error - หลังบ้านพัง Handler ส่งต่อ Error ได้",
			url:  "/api/v1/transaction?month=6&year=2026&page=1&limit=10",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("FetchTransactions", mock.Anything, mockInput).Return(nil, errors.New("something went wrong in usecase"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)

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
			bodyByte, _ := io.ReadAll(resp.Body)
			bodyStr := string(bodyByte)
			assert.Contains(t, bodyStr, tt.expectedBody)

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestUpdateTransaction(t *testing.T) {
	fixedTime := time.Date(2026, time.June, 19, 16, 0, 0, 0, timeutil.BangKokLoc)

	// ข้อมูลเข้า request body ที่ถูกต้อง
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
		CategoryID:      pkg.PTR(int64(3)),
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
			expectedBody:   "Invalid ID format. The path parameter 'id' must be a positive integer.",
		},
		{
			name:        "2. Bad Request - JSON Body รูปแบบไม่ถูกต้อง แกะข้อมูลไม่ได้",
			paramID:     "1",
			requestBody: "{invalid json}",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// usecase จะยังไม่ทำงาน
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Invalid JSON format.",
		},
		{
			name:    "3. Bad Request - ข้อมูลไม่ผ่าน Tag Validation",
			paramID: "1",
			requestBody: dto.UpdateTransactionInput{
				Amount: pkg.PTR(-50.0), // สมมติว่าใน DTO ห้ามติดลบ ทำให้ ValidateStruct พัง
			},
			setupMock:      func(uc *domain.TransactionUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Validation failed for the request data.",
		},

		{
			name:        "4. Internal Server Error - เกิดข้อผิดพลาดที่ Usecase Layer",
			paramID:     "1",
			requestBody: validInput,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("UpdateTransaction", mock.Anything, uint(1), mock.MatchedBy(func(req domain.UpdateTransactionParam) bool {
					// 1. เช็คค่า Pointers ทั่วไป
					if req.Amount == nil || *req.Amount != *validInput.Amount {
						return false
					}
					if req.Note == nil || *req.Note != *validInput.Note {
						return false
					}
					if req.CategoryID == nil || *req.CategoryID != *validInput.CategoryID {
						return false
					}

					// 2. เช็คเวลาโดยใช้ .Equal()
					if !req.TransactionDate.Equal(*validInput.TransactionDate) {
						return false
					}

					return true
				})).Return(nil, domain.ErrInternalDB)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later.",
		},
		{
			name:        "5. Success - อัปเดตข้อมูลสำเร็จ",
			paramID:     "1",
			requestBody: validInput,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("UpdateTransaction", mock.Anything, uint(1), mock.MatchedBy(func(req domain.UpdateTransactionParam) bool {
					// 1. เช็คค่า Pointers ทั่วไป
					if req.Amount == nil || *req.Amount != *validInput.Amount {
						return false
					}
					if req.Note == nil || *req.Note != *validInput.Note {
						return false
					}
					if req.CategoryID == nil || *req.CategoryID != *validInput.CategoryID {
						return false
					}

					// 2. เช็คเวลาโดยใช้ .Equal()
					if !req.TransactionDate.Equal(*validInput.TransactionDate) {
						return false
					}

					return true
				})).Return(mockResult, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
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
			bodyByte, _ := io.ReadAll(resp.Body)
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
			expectedBody:   "Invalid ID format. The path parameter 'id' must be a positive integer.",
		},

		{
			name:    "2. Internal Server Error - เกิดข้อผิดพลาดที่ Usecase Layer",
			paramID: "1",
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				uc.On("DeleteTransaction", mock.Anything, uint(1)).Return(domain.ErrInternalDB)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later.",
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
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
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
			bodyByte, _ := io.ReadAll(resp.Body)
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
			expectedBody:   "Missing required form field 'local_image_name'.",
		},
		{
			name:      "2. Validation Failed - on image file",
			imageName: "slip_06.jpg",
			hasImage:  false,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				// usecase จะไม่ทำงานใน case นี้
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Missing required file 'image' in multipart/form-data request.",
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
			expectedBody:   `"message":"Transaction processed successfully"`,
		},
		{
			name:      "4. Succes - Create Sucess Duplicate Data",
			imageName: "slip_06.jpg",
			hasImage:  true,
			setupMock: func(uc *domain.TransactionUsecaseMock) {
				mockImageByte := []byte("fake-image-bytes")
				uc.On("SyncTransaction", mock.Anything, mockImageByte, "slip_06.jpg").Return(nil, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"message":"Transaction processed successfully (skipped or duplicate caught early)"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockUsecase := new(domain.TransactionUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)

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
