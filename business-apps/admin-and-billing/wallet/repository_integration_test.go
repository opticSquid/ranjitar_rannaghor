package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestProcessRechargeInDB(t *testing.T) {
	testdb.ResetData()
	// create user
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('WRepoUser', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	req := RechargeRequest{UserID: userID, Amount: 50.0, TxnDate: time.Now()}
	newBal, err := ProcessRechargeInDB(context.Background(), req, req.TxnDate)
	require.NoError(t, err)
	require.Equal(t, 50.0, newBal)
}
