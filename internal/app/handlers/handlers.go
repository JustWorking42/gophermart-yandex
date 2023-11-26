package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"

	"github.com/JustWorking42/gophermart-yandex/internal/app/app"
	"github.com/JustWorking42/gophermart-yandex/internal/app/authorization"
	"github.com/JustWorking42/gophermart-yandex/internal/app/cookie"
	"github.com/JustWorking42/gophermart-yandex/internal/app/luna"
	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/JustWorking42/gophermart-yandex/internal/app/model/apperrors"

	"github.com/JustWorking42/gophermart-yandex/internal/app/repository"
	"github.com/go-chi/chi/v5"
)

func Webhooks(app *app.App) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/api/user/register", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization.RegisterHandler(app.Repository, w, r)
	}))

	router.Post("/api/user/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization.LoginHandler(app.Repository, w, r)
	}))

	router.Post("/api/user/orders", cookie.ValidateCookieMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createOrder(app.Repository, w, r)
	})))

	router.Post("/api/user/balance/withdraw", cookie.ValidateCookieMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		withdraw(app.Repository, w, r)
	})))
	router.Get("/api/user/orders", cookie.ValidateCookieMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getOrders(app.Repository, w, r)
	})))
	router.Get("/api/user/balance", cookie.ValidateCookieMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		balance(app.Repository, w, r)
	})))
	router.Get("/api/user/withdrawals", cookie.ValidateCookieMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getWitdrawals(app.Repository, w, r)
	})))

	return router
}

func createOrder(repository *repository.AppRepository, w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(cookie.ContextKeyUsername).(string)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderID := string(body)

	if !luna.Valid(orderID) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = repository.RegisterOrder(r.Context(), model.RegisterOrderModel{
		OrderID: orderID,
	},
		username,
	)

	if err != nil {
		switch err := err.(type) {
		case *apperrors.ErrAlreadyRegisteredByThisUser:
			w.WriteHeader(http.StatusOK)
		case *apperrors.ErrAlreadyRegisteredByAnotherUser:
			http.Error(w, err.Error(), http.StatusConflict)
		case *apperrors.ErrUserDoesNotExist:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func getOrders(repository *repository.AppRepository, w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(cookie.ContextKeyUsername).(string)

	orders, err := repository.GetOrdersByUser(r.Context(), username)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func getWitdrawals(repository *repository.AppRepository, w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(cookie.ContextKeyUsername).(string)

	withdrawals, err := repository.GetWithdrawalsByUser(r.Context(), username)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].ProcessedAt.Before(withdrawals[j].ProcessedAt)
	})

	response, err := json.Marshal(withdrawals)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func withdraw(repository *repository.AppRepository, w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(cookie.ContextKeyUsername).(string)

	var body model.WithdrawalModel
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if !luna.Valid(body.Order) {
		http.Error(w, "Invalid Order Number", http.StatusUnprocessableEntity)
		return
	}

	err := repository.WithdrawBalance(r.Context(), username, body)
	if err != nil {
		if errors.Is(err, &apperrors.ErrInsufficientBalance{}) {
			http.Error(w, "Insufficient Funds", http.StatusPaymentRequired)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func balance(repository *repository.AppRepository, w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(cookie.ContextKeyUsername).(string)

	balance, withdrawn, err := repository.GetBalanceAndWithdrawnInCentsByUser(r.Context(), username)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{
		Current:   float64(balance) / 100.0,
		Withdrawn: float64(withdrawn) / 100.0,
	}

	response, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
