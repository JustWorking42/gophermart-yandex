package authorization

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/JustWorking42/gophermart-yandex/internal/app/model/apperrors"
	"github.com/JustWorking42/gophermart-yandex/internal/app/repository"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte("kety")

func RegisterHandler(repository repository.AppRepository, w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	logger.Sugar().Infow("RegisterHandler started")
	var creds model.RegisterRequestModel
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		logger.Sugar().Errorw("Error decoding request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if creds.Password == "" || creds.Username == "" {
		logger.Sugar().Errorw("Username or password is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := repository.GetByUsername(r.Context(), creds.Username)
	if err != nil {
		if !errors.Is(err, apperrors.ErrUserDoesNotExist) {
			logger.Sugar().Errorw("Error getting user by username", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if user != nil {
		logger.Sugar().Errorw("User alrady Exist", "username", creds.Username)
		w.WriteHeader(http.StatusConflict)
		return
	}

	user, err = generateUser(creds)
	if err != nil {
		logger.Sugar().Errorw("Error generating user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = repository.Register(r.Context(), *user)
	if err != nil {
		logger.Sugar().Errorw("Error registering user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenString, err := generateToken(creds.Username)
	if err != nil {
		logger.Sugar().Errorw("Error generating token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
	})

	logger.Sugar().Infow("RegisterHandler finished successfully", "username", creds.Username)
	w.WriteHeader(http.StatusOK)
}

func LoginHandler(repository repository.AppRepository, w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	logger.Sugar().Infow("LoginHandler started")
	var creds model.RegisterRequestModel
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		logger.Sugar().Errorw("Error decoding request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if creds.Password == "" || creds.Username == "" {
		logger.Sugar().Errorw("Username or password is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := repository.GetByUsername(r.Context(), creds.Username)
	if err != nil {
		if !errors.Is(err, apperrors.ErrUserDoesNotExist) {
			logger.Sugar().Errorw("Error getting user by username", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			logger.Sugar().Errorf("User not found", "username", creds.Username)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(creds.Password))
	if err != nil {
		logger.Sugar().Errorf("Invalid password", "username", creds.Username)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenString, err := generateToken(creds.Username)
	if err != nil {
		logger.Sugar().Errorw("Error generating token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
	})

	logger.Sugar().Infow("LoginHandler finished successfully", "username", creds.Username)
	w.WriteHeader(http.StatusOK)
}

func ParseToken(tokenStr string) (*jwt.MapClaims, error) {
	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func generateUser(data model.RegisterRequestModel) (*model.UserModel, error) {
	var user model.UserModel
	user.Username = data.Username
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.HashedPassword = string(hashedPassword)
	return &user, nil
}

func generateToken(login string) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Username": login,
	})
	tokenString, err = token.SignedString(secretKey)
	if err != nil {
		return
	}
	return
}
