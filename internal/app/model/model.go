package model

import "time"

type RegisterRequestModel struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

type UserModel struct {
	Username       string
	HashedPassword string
	WalletID       int
}

type OrderStatus int

type RegisterOrderModel struct {
	OrderID string
	UserID  int
}

type OrderModel struct {
	OrderID    string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Claims struct {
	Username string
}

type WithdrawalModel struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
