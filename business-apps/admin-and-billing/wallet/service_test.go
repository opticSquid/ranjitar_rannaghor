package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T) int {
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('WServiceUser', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)
	return userID
}

func TestRechargeWalletService_ZeroDateNormalizes(t *testing.T) {
	testdb.ResetData()
	userID := createTestUser(t)

	req := RechargeRequest{UserID: userID, Amount: 123.45}
	// TxnDate is zero value
	newBal, err := RechargeWalletService(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 123.45, newBal)

	// Ensure a transaction exists with created_at close to now
	var createdAt time.Time
	err = testdb.DbPool.QueryRow(context.Background(), `SELECT created_at FROM wallet_transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID).Scan(&createdAt)
	require.NoError(t, err)
	if time.Since(createdAt) > 5*time.Minute {
		t.Fatalf("created_at too old: %v", createdAt)
	}
}
