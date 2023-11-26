package authorization

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/JustWorking42/gophermart-yandex/internal/app/repository"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte("kety")

func RegisterHandler(repository *repository.AppRepository, w http.ResponseWriter, r *http.Request) {
	var creds model.RegisterRequestModel
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if creds.Password == "" || creds.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, _ := repository.GetByUsername(r.Context(), creds.Username)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	if user != model.EmptyUser() {
		w.WriteHeader(http.StatusConflict)
		return
	}

	user, err = generateUser(creds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = repository.Register(r.Context(), user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenString, err := generateToken(creds.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

func LoginHandler(repository *repository.AppRepository, w http.ResponseWriter, r *http.Request) {
	var creds model.RegisterRequestModel
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := repository.GetByUsername(r.Context(), creds.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == model.EmptyUser() {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(creds.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenString, err := generateToken(creds.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
	})

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

func generateUser(data model.RegisterRequestModel) (user model.UserModel, err error) {
	user.Username = data.Username
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	user.HashedPassword = string(hashedPassword)
	return
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
