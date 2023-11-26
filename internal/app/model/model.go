package model

import "time"

type RegisterRequestModel struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

type UserModel struct {
	Username       string
	HashedPassword string
	WalletId       int
}

type OrderStatus int

const (
	NEW OrderStatus = iota
	REGISTERED
	INVALID
	PROCESSING
	PROCESSED
)

type RegisterOrderModel struct {
	OrderID string
	UserID  int
}

type OrderModel struct {
	OrderID    string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Claims struct {
	Username string
}

func EmptyUser() UserModel {
	return UserModel{}
}

type WithdrawalModel struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
