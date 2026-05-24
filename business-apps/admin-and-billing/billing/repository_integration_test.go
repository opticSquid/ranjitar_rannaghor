package billing

import (
	"context"
	"testing"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestFetchBillReportFromDB(t *testing.T) {
	testdb.ResetData()
	// create user
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, mobile_no, building_no, room_no, role, plan) VALUES ('RepoBilling', '', '', '', 'normal', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	// create transactions and logs
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO wallet_transactions (user_id, txn_type, status, amount, balance_after, created_at) VALUES ($1, 'recharge', 'confirmed', 100, 100, '2023-09-30 10:00:00')`, userID)
	require.NoError(t, err)

	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, total_cost, special_dish_name) VALUES ($1, '2023-10-10', 'lunch', true, 52.5, '')`, userID)
	require.NoError(t, err)

	start := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 10, 31, 0, 0, 0, 0, time.UTC)

	report, err := FetchBillReportFromDB(context.Background(), userID, start, end)
	require.NoError(t, err)

	if report.TotalSpent == 0 {
		t.Fatalf("expected total spent > 0")
	}
}
