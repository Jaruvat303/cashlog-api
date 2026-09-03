package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

func TestCreateAccount(t *testing.T) {
	validInput := dto.CreateAccountInput{
		Name:             "Test Account",
		AccountType:      "cash",
		OpeningBalance:   1000.0,
		MatchingKeywords: []string{"test", "account"},
		BankIcon:         "cash",
	}

	mockResult := &domain.Account{
		ID:          1,
		Name:        "Test Account",
		AccountType: "cash",
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(uc *domain.AccountUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - สร้าง Account สำเร็จ",
			requestBody: validInput,
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("CreateAccount", mock.Anything, validInput.ToDomainCreateParam()).Return(mockResult, nil)
			},
			expectedStatus: fiber.StatusCreated,
			expectedBody:   `"success":true`,
		},
		{
			name:           "Bad Request - JSON Body พัง แกะข้อมูลไม่ได้",
			requestBody:    "{ invalid json }",
			setupMock:      func(uc *domain.AccountUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Invalid JSON format.",
		},
		{
			name: "Bad Request - ข้อมูลไม่ผ่าน Validation Struct",
			requestBody: dto.CreateAccountInput{
				Name:        "", // สมมติว่าใน tag บังคับ required ไว้
				AccountType: "", // สมมติว่าใน tag บังคับ required ไว้
			},
			setupMock:      func(uc *domain.AccountUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Validation failed for the request data.",
		},
		{
			name:        "Not Found - Usecase แจ้งว่าหาข้อมูลอ้างอิงไม่เจอ",
			requestBody: validInput,
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("CreateAccount", mock.Anything, validInput.ToDomainCreateParam()).Return(nil, domain.ErrNotFound)
			},
			expectedStatus: fiber.StatusNotFound,
			expectedBody:   "The requested data was not found",
		},
		{
			name:        "Internal Error - Usecase ทำงานผิดพลาด",
			requestBody: validInput,
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("CreateAccount", mock.Anything, validInput.ToDomainCreateParam()).Return(nil, errors.New("db error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(domain.AccountUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
			tt.setupMock(mockUC)

			h := handler.NewAccountHandler(mockUC, mockLog)
			app.Post("/accounts", h.CreateAccount)

			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			assert.Contains(t, bodyString, tt.expectedBody)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	invalidInput := dto.UpdateAccountInput{
		Name:             pkg.PTR("testUpdateName"),
		AccountType:      pkg.PTR("bank"),
		MatchingKeywords: &[]string{"update", "test"},
	}

	mockResult := &domain.Account{
		ID:               1,
		Name:             "testUpdateName",
		AccountType:      "bank",
		MatchingKeywords: []string{"update", "test"},
	}

	tests := []struct {
		name           string
		accountID      string
		requestBody    interface{}
		setupMock      func(uc *domain.AccountUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - อัปเดต Account สำเร็จ",
			accountID:   "1",
			requestBody: invalidInput,
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("UpdateAccount", mock.Anything, uint(1), invalidInput.ToDomainUpdateParam()).Return(mockResult, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name:      "Bad Request - ข้อมูลไม่ผ่าน Validation Struct",
			accountID: "1",
			requestBody: dto.UpdateAccountInput{
				Name:        pkg.PTR(""), // สมมติว่าใน tag บังคับ required ไว้
				AccountType: pkg.PTR(""), // สมมติว่าใน tag บังคับ required ไว้
			},
			setupMock:      func(uc *domain.AccountUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Validation failed for the request data.",
		},
		{
			name:           "Bad Request - accountID ไม่ใช่ตัวเลข",
			accountID:      "abc",
			requestBody:    invalidInput,
			setupMock:      func(uc *domain.AccountUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Invalid ID format. The path parameter 'id' must be a positive integer.",
		},
		{
			name:           "Bad Request - JSON Body พัง แกะข้อมูลไม่ได้",
			accountID:      "1",
			requestBody:    "{ invalid json }",
			setupMock:      func(uc *domain.AccountUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Invalid JSON format.",
		},
		{
			name:        "Not Found - Usecase แจ้งว่าหาข้อมูลไม่เจอหรือลบไม่ได้",
			accountID:   "1",
			requestBody: invalidInput,
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("UpdateAccount", mock.Anything, uint(1), invalidInput.ToDomainUpdateParam()).Return(nil, domain.ErrNotFound)
			},
			expectedStatus: fiber.StatusNotFound,
			expectedBody:   "The requested data was not found",
		},
		{
			name:        "Internal Error - Usecase ทำงานผิดพลาด",
			accountID:   "1",
			requestBody: invalidInput,
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("UpdateAccount", mock.Anything, uint(1), invalidInput.ToDomainUpdateParam()).Return(nil, errors.New("db error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(domain.AccountUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
			tt.setupMock(mockUC)

			h := handler.NewAccountHandler(mockUC, mockLog)
			app.Put("/accounts/:id", h.UpdateAccount)

			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/accounts/"+tt.accountID, bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			assert.Contains(t, bodyString, tt.expectedBody)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		setupMock      func(uc *domain.AccountUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "Success - ลบ Account สำเร็จ",
			accountID: "1",
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("DeleteAccount", mock.Anything, uint(1)).Return(nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name:           "Bad Request - accountID ไม่ใช่ตัวเลข",
			accountID:      "abc",
			setupMock:      func(uc *domain.AccountUsecaseMock) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Invalid ID format. The path parameter 'id' must be a positive integer.",
		},
		{
			name:      "Not Found - Usecase แจ้งว่าหาข้อมูลไม่เจอหรือลบไม่ได้",
			accountID: "1",
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("DeleteAccount", mock.Anything, uint(1)).Return(domain.ErrNotFound)
			},
			expectedStatus: fiber.StatusNotFound,
			expectedBody:   "The requested data was not found",
		},
		{
			name:      "Internal Error - Usecase ทำงานผิดพลาด",
			accountID: "1",
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("DeleteAccount", mock.Anything, uint(1)).Return(errors.New("db error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(domain.AccountUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
			tt.setupMock(mockUC)

			h := handler.NewAccountHandler(mockUC, mockLog)
			app.Delete("/accounts/:id", h.DeleteAccount)

			req := httptest.NewRequest("DELETE", "/accounts/"+tt.accountID, nil)

			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			assert.Contains(t, bodyString, tt.expectedBody)
		})
	}

}

func TestFetchActiveAccounts(t *testing.T) {
	mockAccounts := []domain.Account{
		{ID: 1, Name: "Account 1", AccountType: "cash"},
		{ID: 2, Name: "Account 2", AccountType: "bank"},
	}

	tests := []struct {
		name           string
		setupMock      func(uc *domain.AccountUsecaseMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - ดึงรายการ Account ที่ใช้งานอยู่สำเร็จ",
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("FetchActiveAccounts", mock.Anything).Return(mockAccounts, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name: "Internal Error - Usecase ทำงานผิดพลาด",
			setupMock: func(uc *domain.AccountUsecaseMock) {
				uc.On("FetchActiveAccounts", mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   "Something went wrong, please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(domain.AccountUsecaseMock)
			mockLog := logger.NewNopLogger()
			app := newTestApp(mockLog)
			tt.setupMock(mockUC)

			h := handler.NewAccountHandler(mockUC, mockLog)
			app.Get("/accounts", h.FetchActiveAccounts)

			req := httptest.NewRequest("GET", "/accounts", nil)

			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			assert.Contains(t, bodyString, tt.expectedBody)
		})
	}
}
