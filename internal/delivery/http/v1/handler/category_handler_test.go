package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/handler"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateCategoryHandler(t *testing.T) {
	validInput := dto.CreateCategoryInput{Name: "Food", Type: "expense"}

	mockResult := &domain.Category{ID: 1, Name: "Food", Type: "expense"}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(uc *domain.CategoryUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - สร้าง Category สำเร็จ",
			requestBody: validInput,
			setupMock: func(uc *domain.CategoryUsecaseMock) {
				uc.On("CreateCategory", mock.Anything, validInput.ToDomainCreateParam()).Return(mockResult, nil)
			},
			expectedStatus: fiber.StatusCreated,
			expectedBody:   `"success":true`,
		},
		{
			name:           "Bad Request - JSON Body พัง แกะข้อมูลไม่ได้",
			requestBody:    "{ invalid json }",
			setupMock:      func(uc *domain.CategoryUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Invalid JSON format.",
		},
		{
			name: "Bad Request - ข้อมูลไม่ผ่าน Validation Struct",
			requestBody: dto.CreateCategoryInput{
				Name: "", // สมมติว่าใน tag บังคับ required ไว้
			},
			setupMock:      func(uc *domain.CategoryUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Validation failed for the request data.",
		},
		{
			name:        "Internal Error - Usecase ทำงานผิดพลาด",
			requestBody: validInput,
			setupMock: func(uc *domain.CategoryUsecaseMock) {
				uc.On("CreateCategory", mock.Anything, validInput.ToDomainCreateParam()).Return(nil, errors.New("db error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(domain.CategoryUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
			tt.setupMock(mockUC)

			h := handler.NewCategoryHandler(mockUC, mockLog)
			app.Post("/categories", h.CreateCategory)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			bodyByte, _ := io.ReadAll(resp.Body)
			bodyStr := string(bodyByte)
			assert.Contains(t, bodyStr, tt.expectedBody)

			mockUC.AssertExpectations(t)
		})
	}
}

func TestUpdateCategoryHandler(t *testing.T) {
	validInput := dto.UpdateCategoryInput{Name: pkg.PTR("Shopping")}
	mockResult := &domain.Category{ID: 1, Name: "Shopping"}

	tests := []struct {
		name           string
		paramID        string
		requestBody    interface{}
		setupMock      func(uc *domain.CategoryUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - แก้ไขข้อมูลสำเร็จ",
			paramID:     "1",
			requestBody: validInput,
			setupMock: func(uc *domain.CategoryUsecaseMock) {
				uc.On("UpdateCategory", mock.Anything, uint(1), validInput.ToDomainUpdateParam()).Return(mockResult, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name:           "Bad Request - ID ใน URL พัง ไม่ใช่ตัวเลข",
			paramID:        "xyz",
			requestBody:    validInput,
			setupMock:      func(uc *domain.CategoryUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "parse id param to uint error",
		},
		{
			name:           "Bad Request - JSON Body พัง แกะข้อมูลไม่ได้",
			paramID:        "1",
			requestBody:    "{ invalid json }",
			setupMock:      func(uc *domain.CategoryUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:        "Internal Error - Usecase ทำงานผิดพลาด",
			paramID:     "1",
			requestBody: validInput,
			setupMock: func(uc *domain.CategoryUsecaseMock) {
				uc.On("UpdateCategory", mock.Anything, uint(1), validInput.ToDomainUpdateParam()).Return(nil, errors.New("something went wrong"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(domain.CategoryUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
			tt.setupMock(mockUC)

			h := handler.NewCategoryHandler(mockUC, mockLog)
			app.Patch("/categories/:id", h.UpdateCategory)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPatch, "/categories/"+tt.paramID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestDeleteCategoryHandler(t *testing.T) {
	tests := []struct {
		name           string
		paramID        string
		setupMock      func(uc *domain.CategoryUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success - ลบข้อมูลสำเร็จ",
			paramID: "1",
			setupMock: func(uc *domain.CategoryUsecaseMock) {
				uc.On("DeleteCategory", mock.Anything, uint(1)).Return(nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name:           "Bad Request - ID ใน URL พัง ไม่ใช่ตัวเลข",
			paramID:        "abc",
			setupMock:      func(uc *domain.CategoryUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "parse id param to uint error",
		},
		{
			name:    "Internal Error - Usecase แจ้งว่าหาข้อมูลไม่เจอหรือลบไม่ได้",
			paramID: "1",
			setupMock: func(uc *domain.CategoryUsecaseMock) {
				uc.On("DeleteCategory", mock.Anything, uint(1)).Return(errors.New("not found"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(domain.CategoryUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
			tt.setupMock(mockUC)

			h := handler.NewCategoryHandler(mockUC, mockLog)
			app.Delete("/categories/:id", h.DeleteCategory)

			req := httptest.NewRequest(http.MethodDelete, "/categories/"+tt.paramID, nil)
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestFetchCategoriesByTypeHandler(t *testing.T) {
	mockList := []domain.Category{{ID: 1, Name: "Food", Type: "expense"}}

	tests := []struct {
		name           string
		url            string
		setupMock      func(uc *domain.CategoryUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - ดึงข้อมูลรายการสำเร็จ",
			url:  "/categories?type=expense",
			setupMock: func(uc *domain.CategoryUsecaseMock) {
				uc.On("FetchCategoriesByType", mock.Anything, "expense").Return(mockList, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name: "Bad Request - Query Parameter ไม่ถูกต้อง",
			url:  "/categories?type=wrongType",
			setupMock: func(uc *domain.CategoryUsecaseMock) {
				// usecase จะไม่ทำงาน
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "invalid types parameter",
		},
		{
			name: "Internal Error - Usecase ดึงข้อมูลจากฐานข้อมูลไม่ได้",
			url:  "/categories?type=expense",
			setupMock: func(uc *domain.CategoryUsecaseMock) {
				uc.On("FetchCategoriesByType", mock.Anything, "expense").Return(nil, errors.New("query error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "query error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(domain.CategoryUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
			tt.setupMock(mockUC)

			h := handler.NewCategoryHandler(mockUC, mockLog)
			app.Get("/categories", h.FetchCategoriesByType)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockUC.AssertExpectations(t)
		})
	}
}
