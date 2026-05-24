package utils

type TransactionType string

const (
	RECHARGE TransactionType = "recharge"
	DELIVERY TransactionType = "delivery"
	REFUND   TransactionType = "refund"
)
