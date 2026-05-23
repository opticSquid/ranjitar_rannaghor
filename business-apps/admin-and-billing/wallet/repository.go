package wallet

import (
	"context"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/utils"
)

func ProcessRechargeInDB(ctx context.Context, req RechargeRequest, txnDate time.Time) (float64, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var prevBalanceAfter *float64
	err = tx.QueryRow(ctx, `SELECT BALANCE_AFTER FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 AND CREATED_AT < $2 ORDER BY CREATED_AT DESC LIMIT 1`, req.UserID, txnDate).Scan(&prevBalanceAfter)

	if err != nil && err.Error() != "no rows in result set" {
		return 0, err
	}

	var currentBalance float64 = 0
	if prevBalanceAfter != nil {
		currentBalance = *prevBalanceAfter
	}
	newBalance := currentBalance + req.Amount

	_, err = tx.Exec(ctx, `
		INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, REFERENCE_ID, CREATED_AT)
		VALUES ($1, 'recharge', 'confirmed', $2, $3, $4, $5)
	`, req.UserID, req.Amount, newBalance, req.RefID, txnDate)
	if err != nil {
		return 0, err
	}

	// Recalculate all future balances for this user
	err = utils.RecalculateBalances(ctx, tx, utils.RECHARGE, req.UserID, txnDate, req.Amount)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return newBalance, nil
}
