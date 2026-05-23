package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		INSERT INTO users (name, plan) VALUES ('Wallet User', 'standard') RETURNING user_id
	`).Scan(&userID)
	require.NoError(t, err)
	return userID
}

func TestRechargeWallet_Success(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)

	reqBody := RechargeRequest{
		UserID:  userID,
		Amount:  500.0,
		RefID:   "REF123",
		TxnDate: time.Now(),
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/wallet/recharge", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	RechargeWallet(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]float64
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, 500.0, resp["new_balance"])

	// Check DB
	var count int
	testdb.DbPool.QueryRow(context.Background(), "SELECT count(*) FROM wallet_transactions WHERE user_id = $1", userID).Scan(&count)
	assert.Equal(t, 1, count)
}

func TestRechargeWallet_InvalidJSON(t *testing.T) {
	testdb.ResetData()

	req := httptest.NewRequest("POST", "/wallet/recharge", bytes.NewBuffer([]byte("bad json")))
	rr := httptest.NewRecorder()

	RechargeWallet(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRechargeWallet_MultipleRecharges(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)

	// First recharge
	reqBody1 := RechargeRequest{UserID: userID, Amount: 100.0, TxnDate: time.Now()}
	body1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/wallet/recharge", bytes.NewBuffer(body1))
	rr1 := httptest.NewRecorder()
	RechargeWallet(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Second recharge
	reqBody2 := RechargeRequest{UserID: userID, Amount: 50.0, TxnDate: time.Now().Add(1 * time.Hour)}
	body2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/wallet/recharge", bytes.NewBuffer(body2))
	rr2 := httptest.NewRecorder()
	RechargeWallet(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	var resp map[string]float64
	json.NewDecoder(rr2.Body).Decode(&resp)
	assert.Equal(t, 150.0, resp["new_balance"])
}

func TestRechargeWallet_BackdatedRecharge(t *testing.T) {
	testdb.ResetData()
	userID := createUser(t)

	now := time.Now()

	// 1. Recharge today
	reqBody1 := RechargeRequest{UserID: userID, Amount: 100.0, TxnDate: now}
	body1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/wallet/recharge", bytes.NewBuffer(body1))
	RechargeWallet(httptest.NewRecorder(), req1)

	// 2. Recharge yesterday (backdated)
	reqBody2 := RechargeRequest{UserID: userID, Amount: 200.0, TxnDate: now.Add(-24 * time.Hour)}
	body2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/wallet/recharge", bytes.NewBuffer(body2))
	rr2 := httptest.NewRecorder()
	RechargeWallet(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	// The backdated recharge's immediate new_balance will be 200,
	// but it recalculates the future so the latest transaction should have balance 300.
	var latestBalance float64
	err := testdb.DbPool.QueryRow(context.Background(), `
		SELECT balance_after FROM wallet_transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&latestBalance)
	require.NoError(t, err)

	assert.Equal(t, 300.0, latestBalance)
}
