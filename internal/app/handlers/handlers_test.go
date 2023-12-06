package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JustWorking42/gophermart-yandex/internal/app/cookie"
	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/JustWorking42/gophermart-yandex/internal/app/model/apperrors"
	"github.com/JustWorking42/gophermart-yandex/internal/app/repository"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func TestCreateOrder(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockAppRepository(ctrl)

	tests := []struct {
		name           string
		username       string
		orderID        string
		mockFunc       func()
		expectedStatus int
	}{
		{
			name:     "valid order",
			username: "testuser",
			orderID:  "79927398713",
			mockFunc: func() {
				mockRepo.EXPECT().RegisterOrder(gomock.Any(), model.RegisterOrderModel{
					OrderID: "79927398713",
				}, "testuser").Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:     "order already registered by this user",
			username: "testuser",
			orderID:  "79927398713",
			mockFunc: func() {
				mockRepo.EXPECT().RegisterOrder(gomock.Any(), model.RegisterOrderModel{
					OrderID: "79927398713",
				}, "testuser").Return(apperrors.ErrAlreadyRegisteredByThisUser)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "order already registered by another user",
			username: "testuser",
			orderID:  "79927398713",
			mockFunc: func() {
				mockRepo.EXPECT().RegisterOrder(gomock.Any(), model.RegisterOrderModel{
					OrderID: "79927398713",
				}, "testuser").Return(apperrors.ErrAlreadyRegisteredByAnotherUser)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:     "user does not exist",
			username: "testuser",
			orderID:  "79927398713",
			mockFunc: func() {
				mockRepo.EXPECT().RegisterOrder(gomock.Any(), model.RegisterOrderModel{
					OrderID: "79927398713",
				}, "testuser").Return(apperrors.ErrUserDoesNotExist)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:     "internal server error",
			username: "testuser",
			orderID:  "79927398713",
			mockFunc: func() {
				mockRepo.EXPECT().RegisterOrder(gomock.Any(), model.RegisterOrderModel{
					OrderID: "79927398713",
				}, "testuser").Return(errors.New("internal server error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/orders", bytes.NewBuffer([]byte(tt.orderID)))
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(req.Context(), cookie.ContextKeyUsername, tt.username)
			req = req.WithContext(ctx)

			tt.mockFunc()

			rr := httptest.NewRecorder()
			createOrder(mockRepo, rr, req, logger)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func TestGetOrders(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockAppRepository(ctrl)

	tests := []struct {
		name           string
		username       string
		mockFunc       func()
		expectedStatus int
	}{
		{
			name:     "orders",
			username: "testuser",
			mockFunc: func() {
				mockRepo.EXPECT().GetOrdersByUser(gomock.Any(), "testuser").Return([]model.OrderModel{
					{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "no orders",
			username: "testuser",
			mockFunc: func() {
				mockRepo.EXPECT().GetOrdersByUser(gomock.Any(), "testuser").Return([]model.OrderModel{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:     "internal server error",
			username: "testuser",
			mockFunc: func() {
				mockRepo.EXPECT().GetOrdersByUser(gomock.Any(), "testuser").Return(nil, errors.New("internal server error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/orders", nil)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(req.Context(), cookie.ContextKeyUsername, tt.username)
			req = req.WithContext(ctx)

			tt.mockFunc()

			rr := httptest.NewRecorder()
			getOrders(mockRepo, rr, req, logger)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockAppRepository(ctrl)

	tests := []struct {
		name           string
		username       string
		mockFunc       func()
		expectedStatus int
	}{
		{
			name:     "withdrawals exist",
			username: "testuser",
			mockFunc: func() {
				mockRepo.EXPECT().GetWithdrawalsByUser(gomock.Any(), "testuser").Return([]model.WithdrawalModel{
					{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "no withdrawals",
			username: "testuser",
			mockFunc: func() {
				mockRepo.EXPECT().GetWithdrawalsByUser(gomock.Any(), "testuser").Return([]model.WithdrawalModel{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:     "internal server error",
			username: "testuser",
			mockFunc: func() {
				mockRepo.EXPECT().GetWithdrawalsByUser(gomock.Any(), "testuser").Return(nil, errors.New("internal server error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/withdrawals", nil)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(req.Context(), cookie.ContextKeyUsername, tt.username)
			req = req.WithContext(ctx)

			tt.mockFunc()

			rr := httptest.NewRecorder()
			getWitdrawals(mockRepo, rr, req, logger)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func TestWithdraw(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockAppRepository(ctrl)

	tests := []struct {
		name           string
		username       string
		body           string
		mockFunc       func()
		expectedStatus int
	}{
		{
			name:     "valid withdrawal",
			username: "testuser",
			body:     `{"Order": "79927398713", "Amount": 100}`,
			mockFunc: func() {
				mockRepo.EXPECT().WithdrawBalance(gomock.Any(), "testuser", gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "insufficient funds",
			username: "testuser",
			body:     `{"Order": "79927398713", "Amount": 100}`,
			mockFunc: func() {
				mockRepo.EXPECT().WithdrawBalance(gomock.Any(), "testuser", gomock.Any()).Return(apperrors.ErrInsufficientBalance)
			},
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:     "internal server error",
			username: "testuser",
			body:     `{"Order": "79927398713", "Amount": 100}`,
			mockFunc: func() {
				mockRepo.EXPECT().WithdrawBalance(gomock.Any(), "testuser", gomock.Any()).Return(errors.New("internal server error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/balance/withdraw", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(req.Context(), cookie.ContextKeyUsername, tt.username)
			req = req.WithContext(ctx)

			tt.mockFunc()

			rr := httptest.NewRecorder()
			withdraw(mockRepo, rr, req, logger)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func TestBalance(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockAppRepository(ctrl)

	tests := []struct {
		name           string
		username       string
		mockFunc       func()
		expectedStatus int
	}{
		{
			name:     "valid balance",
			username: "testuser",
			mockFunc: func() {
				mockRepo.EXPECT().GetBalanceAndWithdrawnInCentsByUser(gomock.Any(), "testuser").Return(10000, 5000, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "internal server error",
			username: "testuser",
			mockFunc: func() {
				mockRepo.EXPECT().GetBalanceAndWithdrawnInCentsByUser(gomock.Any(), "testuser").Return(0, 0, errors.New("internal server error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/balance", nil)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(req.Context(), cookie.ContextKeyUsername, tt.username)
			req = req.WithContext(ctx)

			tt.mockFunc()

			rr := httptest.NewRecorder()
			balance(mockRepo, rr, req, logger)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
