package wallet

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func checkUserExist(tx pgx.Tx, ctx context.Context, userId int) (bool, error) {
	var doesExist bool
	err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM PUBLIC.USERS WHERE USER_ID = $1)", userId).Scan(&doesExist)
	if err != nil {
		return false, err
	}
	return doesExist, nil
}

func createTransaction(tx pgx.Tx, ctx context.Context, txn walletTransaction) error {
	res, err := tx.Exec(ctx, `INSERT INTO
		PUBLIC.WALLET_TRANSACTIONS (
			USER_ID,
			TXN_TYPE,
			AMOUNT,
			TXN_TIMESTAMP,
			ORDER_ID
		)
	VALUES
		($1, $2, $3, $4, $5)
	`, txn.userId, txn.txnType, txn.amount, txn.txnTs, txn.orderId)
	if err != nil {
		return err
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("expected to insert 1 row, got %d", res.RowsAffected())
	}
	return nil
}
