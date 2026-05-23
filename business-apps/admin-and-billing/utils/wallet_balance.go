package utils

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Transaction represents a wallet transaction
// Only the fields we need
// Amount: positive for recharge/refund, negative for delivery
// TxnType: 'recharge', 'refund', 'delivery', etc.
type Transaction struct {
	TxnID     int64
	Amount    float64
	TxnType   string
	CreatedAt time.Time
}

// RecalculateBalances recalculates BALANCE_AFTER for all transactions after a given timestamp for a user
func RecalculateBalances(ctx context.Context, tx pgx.Tx, txnType TransactionType, userID int, fromTime time.Time, totalCost float64) error {
	// Update all balance after values depending on the transaction type after fromTime for the given user
	var err error
	if txnType == DELIVERY {
		_, err = tx.Exec(ctx, `UPDATE WALLET_TRANSACTIONS SET BALANCE_AFTER = BALANCE_AFTER - $1 WHERE USER_ID = $2 AND CREATED_AT > $3`, totalCost, userID, fromTime)
	} else {
		_, err = tx.Exec(ctx, `UPDATE WALLET_TRANSACTIONS SET BALANCE_AFTER = BALANCE_AFTER + $1 WHERE USER_ID = $2 AND CREATED_AT > $3`, totalCost, userID, fromTime)
	}
	// if there are no future records present ignore the error
	if err != nil && err.Error() != "no rows in result set" {
		return err
	}
	return nil
}
