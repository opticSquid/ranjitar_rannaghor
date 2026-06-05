package journal

import (
	"context"
	"testing"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestCreateDailyEntryInDB_NoPriorBalance(t *testing.T) {
	testdb.ResetData()
	// create user
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('JRepoUser', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	req := EntryRequest{
		UserID:      userID,
		LogDate:     time.Now().Truncate(24 * time.Hour),
		MealType:    "lunch",
		HasMainMeal: true,
	}

	newBal, err := CreateDailyEntryInDB(context.Background(), req, 52.5, req.LogDate)
	require.NoError(t, err)
	require.Equal(t, -52.5, newBal)
}

func TestCreateDailyEntryInDB_WithPriorRecharge(t *testing.T) {
	testdb.ResetData()
	// create user
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('JRepoUser2', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	logDate := time.Now().Truncate(24 * time.Hour)
	// insert a prior recharge
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO wallet_transactions (user_id, txn_type, amount, balance_after, created_at) VALUES ($1, 'recharge', 100, 100, $2)`, userID, logDate.Add(-1*time.Hour))
	require.NoError(t, err)

	req := EntryRequest{UserID: userID, LogDate: logDate, MealType: "lunch", HasMainMeal: true}
	newBal, err := CreateDailyEntryInDB(context.Background(), req, 52.5, logDate)
	require.NoError(t, err)
	require.Equal(t, 47.5, newBal)
}

func TestDeleteDailyEntryFromDB_Refund(t *testing.T) {
	testdb.ResetData()
	// create user
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('JRepoUser3', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	logDate := time.Now().Truncate(24 * time.Hour)
	// insert a daily log
	var logID int
	err = testdb.DbPool.QueryRow(context.Background(), `INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, total_cost) VALUES ($1, $2, 'lunch', true, 52.5) RETURNING log_id`, userID, logDate).Scan(&logID)
	require.NoError(t, err)

	// insert a delivery txn that made the balance negative
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO wallet_transactions (user_id, txn_type, amount, balance_after, created_at) VALUES ($1, 'delivery', 52.5, -52.5, $2)`, userID, logDate.Add(-1*time.Hour))
	require.NoError(t, err)

	newBal, err := DeleteDailyEntryFromDB(context.Background(), logID)
	require.NoError(t, err)
	require.Equal(t, 0.0, newBal)
}

func TestUpdateDailyEntryInDB_AdjustCost(t *testing.T) {
	testdb.ResetData()
	// create user
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('JRepoUser4', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)
	createdAt := time.Now()
	logDate := createdAt.Truncate(24 * time.Hour)
	// insert a daily log with initial cost
	var logID int
	err = testdb.DbPool.QueryRow(context.Background(), `INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, total_cost) VALUES ($1, $2, 'lunch', true, 52.5) RETURNING log_id`, userID, logDate).Scan(&logID)
	require.NoError(t, err)

	// insert corresponding delivery txn
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO wallet_transactions (user_id, txn_type, amount, balance_after, created_at) VALUES ($1, 'delivery', 52.5, -52.5, $2)`, userID, createdAt)
	require.NoError(t, err)

	// prepare update request to change cost to 120 (special)
	req := EntryRequest{UserID: userID, MealType: "lunch", HasMainMeal: true, IsSpecial: true}
	newBal, err := UpdateDailyEntryInDB(context.Background(), logID, req)
	require.NoError(t, err)
	require.Equal(t, -120.0, newBal)
}

func TestFetchDailyEntries_ByDateAndUser(t *testing.T) {
	testdb.ResetData()
	// create users
	var userID1, userID2 int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('JRepoUser5', 'standard') RETURNING user_id`).Scan(&userID1)
	require.NoError(t, err)
	err = testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, plan) VALUES ('JRepoUser6', 'standard') RETURNING user_id`).Scan(&userID2)
	require.NoError(t, err)

	date, _ := time.Parse("2006-01-02", "2023-10-01")
	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, special_dish_name, total_cost) VALUES ($1, $2, 'lunch', true, '', 52.5), ($3, $2, 'dinner', true, '', 52.5)`, userID1, date, userID2)
	require.NoError(t, err)

	logs, err := FetchDailyEntries(context.Background(), date, userID1)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, "lunch", logs[0].MealType)
}
