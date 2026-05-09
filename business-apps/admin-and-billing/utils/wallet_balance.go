package utils

import (
	"context"
	"database/sql"
	"time"
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
func RecalculateBalances(ctx context.Context, tx *sql.Tx, userID string, fromTime time.Time) error {
	// Fetch all transactions after fromTime, ordered by CREATED_AT, TXN_ID
	rows, err := tx.QueryContext(ctx, `
		SELECT TXN_ID, AMOUNT, TXN_TYPE, CREATED_AT
		FROM WALLET_TRANSACTIONS
		WHERE USER_ID = $1 AND CREATED_AT >= $2
		ORDER BY CREATED_AT ASC, TXN_ID ASC
	`, userID, fromTime)
	if err != nil {
		return err
	}
	defer rows.Close()

	var txns []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.TxnID, &t.Amount, &t.TxnType, &t.CreatedAt); err != nil {
			return err
		}
		txns = append(txns, t)
	}

	// Get the balance just before fromTime
	var prevBalance float64
	var prevBalancePtr *float64
	err = tx.QueryRowContext(ctx, `
		SELECT BALANCE_AFTER FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 AND CREATED_AT < $2 ORDER BY CREATED_AT DESC, TXN_ID DESC LIMIT 1
	`, userID, fromTime).Scan(&prevBalancePtr)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}
	if prevBalancePtr != nil {
		prevBalance = *prevBalancePtr
	} else {
		prevBalance = 0
	}

	// Recalculate balances
	for _, txn := range txns {
		if txn.TxnType == "delivery" {
			prevBalance -= txn.Amount
		} else {
			prevBalance += txn.Amount
		}
		_, err := tx.ExecContext(ctx, `UPDATE WALLET_TRANSACTIONS SET BALANCE_AFTER = $1 WHERE TXN_ID = $2`, prevBalance, txn.TxnID)
		if err != nil {
			return err
		}
	}
	return nil
}
