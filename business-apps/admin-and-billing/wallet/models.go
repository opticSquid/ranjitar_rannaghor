package wallet

import (
	"errors"
	"time"
)

var (
	ErrUserDoesNotExist = errors.New("user does not exist")
)

type RechargeRequest struct {
	UserID  int       `json:"user_id"`
	Amount  float64   `json:"amount"`
	RefID   string    `json:"ref_id"`
	TxnDate time.Time `json:"txn_date"`
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
