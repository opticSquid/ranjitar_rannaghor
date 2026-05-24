package stats

import (
	"context"
	"testing"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestNetProfitAndProfitPercentage(t *testing.T) {
	testdb.ResetData()
	// create a user
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('StatsUser', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	// Add a revenue (recharge)
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO wallet_transactions (user_id, txn_type, status, amount, balance_after, created_at) VALUES ($1, 'recharge', 'confirmed', 200, 200, $2)`, userID, time.Now())
	require.NoError(t, err)

	// Add an expense (daily_logs) as spent
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, total_cost) VALUES ($1, $2, 'lunch', true, 50)`, userID, time.Now())
	require.NoError(t, err)

	// Dashboard stats
	stats, err := GetDashboardStatsService(context.Background())
	require.NoError(t, err)
	require.Equal(t, stats.TotalRevenue-stats.TotalExpenses, stats.NetProfit)

	// Analytics stats (profit percentage)
	astats, err := GetAnalyticsStatsService(context.Background())
	require.NoError(t, err)
	if astats.TotalRevenue > 0 {
		require.Equal(t, ((astats.TotalRevenue-astats.TotalExpenses)/astats.TotalRevenue)*100, astats.ProfitPercentage)
	}
}
