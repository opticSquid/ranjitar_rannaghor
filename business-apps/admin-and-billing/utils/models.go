package utils

type TransactionType int

const (
	RECHARGE TransactionType = iota
	DELIVERY
	REFUND
)
