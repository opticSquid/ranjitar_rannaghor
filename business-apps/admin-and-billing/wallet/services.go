package wallet

import (
	"context"
	"time"
)

func RechargeWalletService(ctx context.Context, req RechargeRequest) (float64, error) {
	txnDate := req.TxnDate
	if txnDate.IsZero() {
		txnDate = time.Now()
	}
	return ProcessRechargeInDB(ctx, req, txnDate)
}
