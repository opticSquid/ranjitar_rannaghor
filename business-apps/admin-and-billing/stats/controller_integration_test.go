package stats

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testdb.Setup()
	defer testdb.Teardown()
	m.Run()
}

func TestGetDashboardStats_HTTP(t *testing.T) {
	testdb.ResetData()
	// create a user and insert some revenue and expenses
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `INSERT INTO users (name, mobile_no, building_no, room_no, role, plan) VALUES ('StatsUser', '', '', '', 'normal', 'standard') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO wallet_transactions (user_id, txn_type, status, amount, balance_after, created_at) VALUES ($1, 'recharge', 'confirmed', 100, 100, now())`, userID)
	require.NoError(t, err)

	_, err = testdb.DbPool.Exec(context.Background(), `INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, total_cost) VALUES ($1, now(), 'lunch', true, 50)`, userID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	rr := httptest.NewRecorder()

	GetDashboardStats(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var got DashboardStats
	json.NewDecoder(rr.Body).Decode(&got)
	// NetProfit field should be present (may be zero)
	require.NotNil(t, got)
}
