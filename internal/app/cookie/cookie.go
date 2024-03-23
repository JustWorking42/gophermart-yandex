package cookie

import (
	"context"
	"errors"
	"net/http"

	"github.com/JustWorking42/gophermart-yandex/internal/app/authorization"
	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const ContextKeyUsername contextKey = "username"

func ValidateCookieMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenStr := cookie.Value

		claims, err := authorization.ParseToken(tokenStr)
		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUsername, (*claims)["Username"])
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
