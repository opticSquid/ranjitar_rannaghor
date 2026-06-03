package journal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
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
	t.Helper()
	var userID int
	err := testdb.DbPool.QueryRow(context.Background(), `
		INSERT INTO users (name, plan) VALUES ('Journal User', 'standard') RETURNING user_id
	`).Scan(&userID)
	require.NoError(t, err)
	return userID
}

func TestCreateDailyEntry_Success(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)

	reqBody := EntryRequest{
		UserID:      userID,
		LogDate:     time.Now().Truncate(24 * time.Hour),
		MealType:    "lunch",
		HasMainMeal: true,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/daily-entry", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	CreateDailyEntry(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp map[string]float64
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, -52.5, resp["new_balance"])
}

func TestCreateDailyEntry_DeductsFromWallet(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)
	logDate := time.Now().Truncate(24 * time.Hour)

	_, err := testdb.DbPool.Exec(context.Background(), `
		INSERT INTO wallet_transactions (user_id, txn_type, amount, balance_after, created_at)
		VALUES ($1, 'recharge', 100, 100, $2)
	`, userID, logDate.Add(-1*time.Hour))
	require.NoError(t, err)

	reqBody := EntryRequest{
		UserID:      userID,
		LogDate:     logDate,
		MealType:    "lunch",
		HasMainMeal: true,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/daily-entry", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	CreateDailyEntry(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp map[string]float64
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, 47.5, resp["new_balance"]) // 100 - 52.5
}

func TestDeleteDailyEntry_Refund(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)
	logDate := time.Now().Truncate(24 * time.Hour)

	var logID int
	err := testdb.DbPool.QueryRow(context.Background(), `
		INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, total_cost)
		VALUES ($1, $2, 'lunch', true, 52.5) RETURNING log_id
	`, userID, logDate).Scan(&logID)
	require.NoError(t, err)

	_, err = testdb.DbPool.Exec(context.Background(), `
		INSERT INTO wallet_transactions (user_id, txn_type, amount, balance_after, created_at)
		VALUES ($1, 'delivery', 52.5, -52.5, $2)
	`, userID, logDate.Add(-1*time.Hour))
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Delete("/daily-entry/{id}", DeleteDailyEntry)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/daily-entry/%d", logID), nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]float64
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, 0.0, resp["new_balance"]) // -52.5 + 52.5 refund
}

func TestUpdateDailyEntry_CostDiff(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)
	createdAt := time.Now()
	logDate := createdAt.Truncate(24 * time.Hour)

	var logID int
	err := testdb.DbPool.QueryRow(context.Background(), `
		INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, total_cost)
		VALUES ($1, $2, 'lunch', true, 52.5) RETURNING log_id
	`, userID, logDate).Scan(&logID)
	require.NoError(t, err)

	_, err = testdb.DbPool.Exec(context.Background(), `
		INSERT INTO wallet_transactions (user_id, txn_type, amount, balance_after, created_at)
		VALUES ($1, 'delivery', 52.5, -52.5, $2)
	`, userID, createdAt)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Put("/daily-entry/{id}", UpdateDailyEntry)

	reqBody := EntryRequest{
		UserID:      userID,
		MealType:    "lunch",
		HasMainMeal: true,
		IsSpecial:   true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/daily-entry/%d", logID), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]float64
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, -120.0, resp["new_balance"])
}

func TestUpdateDailyEntry_NoCostChange(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)
	logDate := time.Now().Truncate(24 * time.Hour)

	var logID int
	err := testdb.DbPool.QueryRow(context.Background(), `
		INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, total_cost)
		VALUES ($1, $2, 'lunch', true, 52.5) RETURNING log_id
	`, userID, logDate).Scan(&logID)
	require.NoError(t, err)

	_, err = testdb.DbPool.Exec(context.Background(), `
		INSERT INTO wallet_transactions (user_id, txn_type, amount, balance_after, created_at)
		VALUES ($1, 'delivery', 52.5, -52.5, $2)
	`, userID, logDate)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Put("/daily-entry/{id}", UpdateDailyEntry)

	reqBody := EntryRequest{UserID: userID, MealType: "dinner", HasMainMeal: true}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/daily-entry/%d", logID), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]float64
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, -52.5, resp["new_balance"]) // No change
}

func TestGetDailyEntries_ByDate(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)
	date, _ := time.Parse("2006-01-02", "2023-10-01")

	_, err := testdb.DbPool.Exec(context.Background(), `
		INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, special_dish_name, total_cost)
		VALUES ($1, $2, 'lunch', true, '', 52.5)
	`, userID, date)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/daily-entries?date=2023-10-01", nil)
	rr := httptest.NewRecorder()
	GetDailyEntries(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var logs []DailyLog
	json.NewDecoder(rr.Body).Decode(&logs)
	require.Len(t, logs, 1)
	assert.Equal(t, "lunch", logs[0].MealType)
}

func TestGetDailyEntries_ByUser(t *testing.T) {
	testdb.ResetData()
	userID1 := createUser(t)
	userID2 := createUser(t)
	date, _ := time.Parse("2006-01-02", "2023-10-01")

	_, err := testdb.DbPool.Exec(context.Background(), `
		INSERT INTO daily_logs (user_id, log_date, meal_type, has_main_meal, special_dish_name, total_cost)
		VALUES ($1, $2, 'lunch', true, '', 52.5), ($3, $2, 'dinner', true, '', 52.5)
	`, userID1, date, userID2)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", fmt.Sprintf("/daily-entries?date=2023-10-01&user_id=%d", userID1), nil)
	rr := httptest.NewRecorder()
	GetDailyEntries(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var logs []DailyLog
	json.NewDecoder(rr.Body).Decode(&logs)
	require.Len(t, logs, 1)
	assert.Equal(t, userID1, logs[0].UserID)
}

func TestGetDailyEntries_MissingDate(t *testing.T) {
	testdb.ResetData()
	req := httptest.NewRequest("GET", "/daily-entries", nil)
	rr := httptest.NewRecorder()
	GetDailyEntries(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
