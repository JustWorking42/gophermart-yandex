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
	"go.uber.org/zap"

	"github.com/JustWorking42/gophermart-yandex/internal/app/repository"
	"github.com/go-chi/chi/v5"
)

func Webhooks(app *app.App) *chi.Mux {
	router := chi.NewRouter()

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			authorization.RegisterHandler(app.Repository, w, r, app.Logger)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			authorization.LoginHandler(app.Repository, w, r, app.Logger)
		})

		r.With(cookie.ValidateCookieMiddleware).Group(func(r chi.Router) {
			r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
				createOrder(app.Repository, w, r, app.Logger)
			})

			r.Post("/balance/withdraw", func(w http.ResponseWriter, r *http.Request) {
				withdraw(app.Repository, w, r, app.Logger)
			})

			r.Get("/orders", func(w http.ResponseWriter, r *http.Request) {
				getOrders(app.Repository, w, r, app.Logger)
			})

			r.Get("/balance", func(w http.ResponseWriter, r *http.Request) {
				balance(app.Repository, w, r, app.Logger)
			})

			r.Get("/withdrawals", func(w http.ResponseWriter, r *http.Request) {
				getWitdrawals(app.Repository, w, r, app.Logger)
			})
		})
	})

	return router
}
func createOrder(repository repository.AppRepository, w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	logger.Sugar().Infof("createOrder started")
	username := r.Context().Value(cookie.ContextKeyUsername).(string)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Sugar().Errorf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderID := string(body)

	if !luna.Valid(orderID) {
		logger.Sugar().Errorf("Invalid order ID: %s", orderID)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = repository.RegisterOrder(r.Context(), model.RegisterOrderModel{
		OrderID: orderID,
	},
		username,
	)

	if err != nil {
		switch err {
		case apperrors.ErrAlreadyRegisteredByThisUser:
			logger.Sugar().Infof("Order already registered by this user: %s", username)
			w.WriteHeader(http.StatusOK)
		case apperrors.ErrAlreadyRegisteredByAnotherUser:
			logger.Sugar().Errorf("Order already registered by another user: %v", err.Error())
			http.Error(w, err.Error(), http.StatusConflict)
		case apperrors.ErrUserDoesNotExist:
			logger.Sugar().Errorf("User does not exist: %v", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			logger.Sugar().Errorf("Internal Server Error: %v", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	logger.Sugar().Infof("Order created successfully: %s", orderID)
	w.WriteHeader(http.StatusAccepted)
}

func getOrders(repository repository.AppRepository, w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	logger.Sugar().Infof("getOrders started")
	username := r.Context().Value(cookie.ContextKeyUsername).(string)

	orders, err := repository.GetOrdersByUser(r.Context(), username)
	if err != nil {
		logger.Sugar().Errorf("Error getting orders by user: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		logger.Sugar().Infof("No orders found for user: %s", username)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		logger.Sugar().Errorf("Error marshalling orders: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Sugar().Infof("Successfully retrieved orders for user: %s", username)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func getWitdrawals(repository repository.AppRepository, w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	logger.Sugar().Infof("getWitdrawals started")
	username := r.Context().Value(cookie.ContextKeyUsername).(string)

	withdrawals, err := repository.GetWithdrawalsByUser(r.Context(), username)
	if err != nil {
		logger.Sugar().Errorf("Error getting withdrawals by user: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		logger.Sugar().Infof("No withdrawals found for user: %s", username)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].ProcessedAt.Before(withdrawals[j].ProcessedAt)
	})

	response, err := json.Marshal(withdrawals)
	if err != nil {
		logger.Sugar().Errorf("Error marshalling withdrawals: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Sugar().Infof("Successfully retrieved withdrawals for user: %s", username)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func withdraw(repository repository.AppRepository, w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	logger.Sugar().Infof("withdraw started")
	username := r.Context().Value(cookie.ContextKeyUsername).(string)

	var body model.WithdrawalModel
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Sugar().Errorf("Error decoding request body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if !luna.Valid(body.Order) {
		logger.Sugar().Errorf("Invalid Order Number: %s", body.Order)
		http.Error(w, "Invalid Order Number", http.StatusUnprocessableEntity)
		return
	}

	err := repository.WithdrawBalance(r.Context(), username, body)
	if err != nil {
		if errors.Is(err, apperrors.ErrInsufficientBalance) {
			logger.Sugar().Errorf("Insufficient Funds for user: %s", username)
			http.Error(w, "Insufficient Funds", http.StatusPaymentRequired)
		} else {
			logger.Sugar().Errorf("Internal Server Error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	logger.Sugar().Infof("Withdrawal successful for user: %s", username)
	w.WriteHeader(http.StatusOK)
}

func balance(repository repository.AppRepository, w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	logger.Sugar().Infof("balance started")
	username := r.Context().Value(cookie.ContextKeyUsername).(string)

	balance, withdrawn, err := repository.GetBalanceAndWithdrawnInCentsByUser(r.Context(), username)
	if err != nil {
		logger.Sugar().Errorf("Error getting balance and withdrawn amount: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := model.Balance{
		Current:   float64(balance) / 100.0,
		Withdrawn: float64(withdrawn) / 100.0,
	}

	response, err := json.Marshal(res)
	if err != nil {
		logger.Sugar().Errorf("Error marshalling balance and withdrawn amount: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Sugar().Infof("Successfully retrieved balance and withdrawn amount for user: %s", username)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
