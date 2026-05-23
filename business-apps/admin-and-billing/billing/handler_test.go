package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testdb.Setup()
	defer testdb.Teardown()
	m.Run()
}

func createUser(t *testing.T) int {
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `
		INSERT INTO users (name, mobile_no, building_no, room_no, role, plan)
		VALUES ('Billing User', '', '', '', 'normal', 'standard') RETURNING user_id
	`).Scan(&userID)
	require.NoError(t, err)
	return userID
}

func TestGetBill_FullReport(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)

	startDate := "2023-10-01"
	endDate := "2023-10-31"

	// Insert recharge (txn before start date to act as opening balance)
	_, err := testdb.DbPool.Exec(context.Background(), `
		INSERT INTO wallet_transactions (user_id, txn_type, status, amount, balance_after, created_at)
		VALUES ($1, 'recharge', 'confirmed', 100, 100, '2023-09-30 10:00:00')
	`, userID)
	require.NoError(t, err)

	// Recharge during period
	_, err = testdb.DbPool.Exec(context.Background(), `
		INSERT INTO wallet_transactions (user_id, txn_type, status, amount, balance_after, created_at)
		VALUES ($1, 'recharge', 'confirmed', 50, 150, '2023-10-05 10:00:00')
	`, userID)
	require.NoError(t, err)

	// Delivery during period
	_, err = testdb.DbPool.Exec(context.Background(), `
		INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, total_cost, special_dish_name)
		VALUES ($1, '2023-10-10', 'lunch', true, 52.5, '')
	`, userID)
	require.NoError(t, err)

	_, err = testdb.DbPool.Exec(context.Background(), `
		INSERT INTO wallet_transactions (user_id, txn_type, status, amount, balance_after, created_at)
		VALUES ($1, 'delivery', 'confirmed', 52.5, 97.5, '2023-10-10 10:00:00')
	`, userID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", fmt.Sprintf("/bill?user_id=%d&start_date=%s&end_date=%s", userID, startDate, endDate), nil)
	rr := httptest.NewRecorder()

	GetBill(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var report BillReport
	json.NewDecoder(rr.Body).Decode(&report)

	assert.Equal(t, 100.0, report.OpeningBalance)
	assert.Equal(t, 50.0, report.TotalRecharges)
	assert.Equal(t, 52.5, report.TotalSpent)
	assert.Equal(t, 97.5, report.ClosingBalance)
	require.Len(t, report.Logs, 1)
}

func TestGetBill_EmptyPeriod(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)

	startDate := "2023-10-01"
	endDate := "2023-10-31"

	req := httptest.NewRequest("GET", fmt.Sprintf("/bill?user_id=%d&start_date=%s&end_date=%s", userID, startDate, endDate), nil)
	rr := httptest.NewRecorder()

	GetBill(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var report BillReport
	json.NewDecoder(rr.Body).Decode(&report)

	assert.Equal(t, 0.0, report.OpeningBalance)
	assert.Equal(t, 0.0, report.TotalRecharges)
	assert.Equal(t, 0.0, report.TotalSpent)
	assert.Empty(t, report.Logs)
}

func TestGetBill_OpeningBalance(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)

	// Insert txn way before
	_, err := testdb.DbPool.Exec(context.Background(), `
		INSERT INTO wallet_transactions (user_id, txn_type, status, amount, balance_after, created_at)
		VALUES ($1, 'recharge', 'confirmed', 123.45, 123.45, '2023-01-01 10:00:00')
	`, userID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", fmt.Sprintf("/bill?user_id=%d&start_date=2023-10-01&end_date=2023-10-31", userID), nil)
	rr := httptest.NewRecorder()

	GetBill(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var report BillReport
	json.NewDecoder(rr.Body).Decode(&report)

	assert.Equal(t, 123.45, report.OpeningBalance)
}
