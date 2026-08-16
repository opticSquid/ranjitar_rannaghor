package orderbook

import (
	"errors"
	"time"
)

var (
	ErrUserDoesNotExist     = errors.New("user does not exist")
	ErrMenuItemDoesNotExist = errors.New("menu item does not exist")
	ErrInvalidMealType      = errors.New("invalid meal type")
	ErrOrderDoesNotExist    = errors.New("order does not exist")
)

type MealType string

const (
	LUNCH  MealType = "lunch"
	DINNER MealType = "dinner"
)

type OrderDetailsRequest struct {
	ItemId   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type NewOrderRequest struct {
	UserId       int                   `json:"user_id"`
	EntryDate    time.Time             `json:"log_date"`
	MealType     MealType              `json:"meal_type"`
	OrderDetails []OrderDetailsRequest `json:"order_details"`
}

type NewOrderResponse struct {
	OrderId int         `json:"order_id"`
	Status  OrderStatus `json:"status"`
}

type OrderStatus string

const (
	COMPLETED OrderStatus = "completed"
	PENDING   OrderStatus = "pending"
	CANCELLED OrderStatus = "cancelled"
)

type order struct {
	orderId     int
	userId      int
	orderTs     time.Time
	mealType    MealType
	totalAmount float64
	status      OrderStatus
}

type orderItem struct {
	orderItemId int
	orderId     int
	itemId      int
	priceId     int
	quantity    int
	subTotal    float64
}

type TxnType string

const (
	RECHARGE TxnType = "recharge"
	DELIVERY TxnType = "delivery"
	REFUND   TxnType = "refund"
)

type walletTransaction struct {
	txnId       int
	userId      int
	orderId     int
	txnType     TxnType
	amount      float64
	txnTs       time.Time
	referenceId string
}

type price struct {
	priceId int
	price   float64
}
