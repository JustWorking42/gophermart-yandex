package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateCookieMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "valid token",
			token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VybmFtZSI6IkJvcmsyIn0.53ea0hPRnOYBTmAJZTauTPopbQwFIT0J87UqcXr9VTM",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid token",
			token:          "invalidtoken",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "no token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.token != "" {
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: tt.token,
				})
			}

			rr := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				username := r.Context().Value(ContextKeyUsername)
				if username != "Bork2" {
					t.Errorf("Username in context is wrong: got %v want %v", username, "Bork2")
				}
			})

			handler := ValidateCookieMiddleware(nextHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}
