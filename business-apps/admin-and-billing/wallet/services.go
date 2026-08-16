package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func rechargeWalletService(ctx context.Context, r RechargeRequest) error {
	if r.TxnDate.IsZero() {
		r.TxnDate = time.Now()
	}
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	doesExist, err := checkUserExist(tx, ctx, r.UserID)
	if err != nil {
		return fmt.Errorf("failed to check user exist: %w", err)
	}
	if !doesExist {
		return fmt.Errorf("user does not exist. %w", ErrUserDoesNotExist)
	}
	err = createTransaction(tx, ctx, walletTransaction{
		userId:      r.UserID,
		txnType:     RECHARGE,
		amount:      r.Amount,
		txnTs:       r.TxnDate,
		referenceId: r.RefID,
	})
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
