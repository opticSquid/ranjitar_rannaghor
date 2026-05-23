package wallet

import "time"

type RechargeRequest struct {
	UserID  int       `json:"user_id"`
	Amount  float64   `json:"amount"`
	RefID   string    `json:"ref_id"`
	TxnDate time.Time `json:"txn_date"`
}
