package authorization

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/JustWorking42/gophermart-yandex/internal/app/model/apperrors"
	"github.com/JustWorking42/gophermart-yandex/internal/app/repository"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func TestRegisterHandler(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockAppRepository(ctrl)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		mockFunc       func()
	}{
		{
			name:           "valid registration",
			body:           `{"login": "testuser", "password": "testpass"}`,
			expectedStatus: http.StatusOK,
			mockFunc: func() {
				mockRepo.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(nil, apperrors.ErrUserDoesNotExist)
				mockRepo.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:           "empty login",
			body:           `{"login": "", "password": "testpass"}`,
			expectedStatus: http.StatusBadRequest,
			mockFunc:       func() {},
		},
		{
			name:           "empty password",
			body:           `{"login": "testuser", "password": ""}`,
			expectedStatus: http.StatusBadRequest,
			mockFunc:       func() {},
		},
		{
			name:           "user alredy exist",
			body:           `{"login": "testuser", "password": "testpass"}`,
			expectedStatus: http.StatusConflict,
			mockFunc: func() {
				user := model.UserModel{}
				mockRepo.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(&user, nil)
			},
		},
		{
			name:           "server error",
			body:           `{"login": "testuser", "password": "testpass"}`,
			expectedStatus: http.StatusInternalServerError,
			mockFunc: func() {
				mockRepo.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(nil, errors.New(""))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/register", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			tt.mockFunc()

			rr := httptest.NewRecorder()
			RegisterHandler(mockRepo, rr, req, logger)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockAppRepository(ctrl)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		mockFunc       func()
	}{
		{
			name:           "valid login",
			body:           `{"login": "testuser", "password": "testpass"}`,
			expectedStatus: http.StatusOK,
			mockFunc: func() {
				user := model.UserModel{HashedPassword: "$2a$10$Xl20X8ZBZuU7uEgoPYO1/uQKY2k3t4vhWOfCnX2ZPOIizMcgr0UbW"} // bcrypt hash of "testpass"
				mockRepo.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(&user, nil)
			},
		},
		{
			name:           "invalid password",
			body:           `{"login": "testuser", "password": "wrongpass"}`,
			expectedStatus: http.StatusUnauthorized,
			mockFunc: func() {
				user := model.UserModel{HashedPassword: "@2a$10$Xl20X8ZBZuU7uEgoPYO1/uQKY2k3t4vhWOfCnX2ZPOIizMcgr0UbW"} // bcrypt hash of "testpass"
				mockRepo.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(&user, nil)
			},
		},
		{
			name:           "user does not exist",
			body:           `{"login": "nonexistent", "password": "testpass"}`,
			expectedStatus: http.StatusUnauthorized,
			mockFunc: func() {
				mockRepo.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(nil, apperrors.ErrUserDoesNotExist)
			},
		},
		{
			name:           "invalid request",
			body:           `{"password": "testpass"}`,
			expectedStatus: http.StatusBadRequest,
			mockFunc:       func() {},
		},
		{
			name:           "server error",
			body:           `{"login": "testuser", "password": "testpass"}`,
			expectedStatus: http.StatusInternalServerError,
			mockFunc: func() {
				mockRepo.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(nil, errors.New(""))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/login", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			tt.mockFunc()

			rr := httptest.NewRecorder()
			LoginHandler(mockRepo, rr, req, logger)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VybmFtZSI6IkJvcmsyIn0.53ea0hPRnOYBTmAJZTauTPopbQwFIT0J87UqcXr9VTM",
			wantErr: false,
		},
		{
			name:    "invalid token",
			token:   "invalidtoken",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
