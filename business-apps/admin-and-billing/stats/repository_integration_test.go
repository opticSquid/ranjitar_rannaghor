package stats

import (
	"context"
	"testing"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestFetchDashboardStatsFromDB(t *testing.T) {
	testdb.ResetData()
	// seed data
	// users
	var userID1, userID2 int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('SUser1', 'standard') RETURNING user_id`).Scan(&userID1)
	require.NoError(t, err)
	err = testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('SUser2', 'standard') RETURNING user_id`).Scan(&userID2)
	require.NoError(t, err)

	// daily_logs revenue
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, special_dish_name, total_cost) VALUES ($1, '2023-10-01', 'lunch', true, '', 50.0), ($2, '2023-10-02', 'dinner', true, '', 30.0)`, userID1, userID2)
	require.NoError(t, err)

	// expenses
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO expenses (expense_date, reason, amount) VALUES ('2023-10-01', 'Fuel', 20.0)`)
	require.NoError(t, err)

	// wallet confirmed balances
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO wallet_transactions (user_id, txn_type, status, amount, balance_after, created_at) VALUES ($1, 'recharge', 'confirmed', 100, 100, '2023-10-01'), ($2, 'recharge', 'confirmed', 50, 50, '2023-10-02')`, userID1, userID2)
	require.NoError(t, err)

	var s DashboardStats
	firstOfMonth, _ := time.Parse("2006-01-02", "2023-10-01")
	err = FetchDashboardStatsFromDB(context.Background(), &s, firstOfMonth)
	require.NoError(t, err)

	// total revenue should be 80 (50 + 30)
	require.Equal(t, 80.0, s.TotalRevenue)
	// total expenses 20
	require.Equal(t, 20.0, s.TotalExpenses)
	// active customers 2
	require.Equal(t, 2, s.ActiveCustomers)
	// wallet pool 150 (sum of latest balances)
	require.Equal(t, 150.0, s.WalletPool)
}

func TestFetchAnalyticsStatsFromDB(t *testing.T) {
	testdb.ResetData()
	// seed daily_logs across multiple dates
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('AUser', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, is_special, special_dish_name, total_cost) VALUES ($1, '2023-10-01', 'lunch', true, false, '', 50.0), ($1, '2023-10-02', 'dinner', true, true, 'SpecialDish', 120.0)`, userID)
	require.NoError(t, err)

	// expenses
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO expenses (expense_date, reason, amount) VALUES ('2023-10-01', 'Fuel', 20.0), ('2023-10-02', 'Supplies', 10.0)`)
	require.NoError(t, err)

	var a AnalyticsStats
	startDate, _ := time.Parse("2006-01-02", "2023-10-01")
	err = FetchAnalyticsStatsFromDB(context.Background(), &a, startDate)
	require.NoError(t, err)

	// trends length should be 31 days (0..30)
	require.Equal(t, 31, len(a.Trends))
	// total revenue should equal 170 (50 + 120)
	require.Equal(t, 170.0, a.TotalRevenue)
	// total expenses should equal 30 (20 + 10)
	require.Equal(t, 30.0, a.TotalExpenses)
	// meal types counts
	require.Equal(t, 1, a.MealTypes["Standard"]) // one non-special
	require.Equal(t, 1, a.MealTypes["Special"])  // one special
	// shifts
	require.Equal(t, 1, a.Shifts["Lunch"])
	require.Equal(t, 1, a.Shifts["Dinner"])
}
