package reports

import (
	"errors"
	"time"
)

var (
	ErrUserDoesNotExist = errors.New("user does not exist")
)

type GenerateBillRequest struct {
	UserId    int       `json:"user_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type MealType string

const (
	LUNCH  MealType = "lunch"
	DINNER MealType = "dinner"
)

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

type OrderItems struct {
	ItemId   int    `json:"item_id"`
	ItemName string `json:"item_name"`
	Quantity int    `json:"quantity"`
}
type Delivery struct {
	OrderId    int          `json:"order_id"`
	TxnDate    time.Time    `json:"txn_date"`
	MealType   MealType     `json:"meal_type"`
	Price      float64      `json:"price"`
	OrderItems []OrderItems `json:"order_items"`
}

type GenerateBillResponse struct {
	UserId                 int        `json:"user_id"`
	UserName               string     `json:"user_name"`
	StartDate              time.Time  `json:"start_date"`
	EndDate                time.Time  `json:"end_date"`
	Balance                float64    `json:"balance"`
	PrevStartingDayBalance float64    `json:"prev_start_day_balance"`
	Deliveries             []Delivery `json:"deliveries"`
}
