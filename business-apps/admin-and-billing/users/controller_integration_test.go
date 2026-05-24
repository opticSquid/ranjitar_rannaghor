package users

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

func TestCreateUser_Success(t *testing.T) {
	testdb.ResetData()

	u := User{
		Name:       "Test User",
		MobileNo:   "1234567890",
		BuildingNo: "A1",
		RoomNo:     "101",
		Role:       "normal",
		Plan:       "standard",
	}
	body, _ := json.Marshal(u)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	CreateUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp User
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.True(t, resp.UserID > 0)
	assert.Equal(t, "Test User", resp.Name)
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	testdb.ResetData()

	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer([]byte("invalid json")))
	rr := httptest.NewRecorder()

	CreateUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateUser_DefaultRole(t *testing.T) {
	testdb.ResetData()

	u := User{
		Name: "Test User 2",
		Plan: "standard",
	}
	body, _ := json.Marshal(u)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	CreateUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp User
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "normal", resp.Role)
}

func TestGetUsers_Empty(t *testing.T) {
	testdb.ResetData()

	req := httptest.NewRequest("GET", "/users", nil)
	rr := httptest.NewRecorder()

	GetUsers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var users []User
	err := json.NewDecoder(rr.Body).Decode(&users)
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestGetUsers_WithBalance(t *testing.T) {
	testdb.ResetData()

	// 1. Create a user
	u := User{Name: "Bal User", Plan: "standard"}
	body, _ := json.Marshal(u)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	CreateUser(rr, req)
	var resp User
	json.NewDecoder(rr.Body).Decode(&resp)

	// 2. Insert wallet transaction
	_, err := testdb.DbPool.Exec(context.Background(), `
		INSERT INTO wallet_transactions (user_id, txn_type, amount, balance_after, created_at)
		VALUES ($1, 'recharge', 100, 150.50, $2)
	`, resp.UserID, time.Now())
	require.NoError(t, err)

	// 3. Get users
	req2 := httptest.NewRequest("GET", "/users", nil)
	rr2 := httptest.NewRecorder()
	GetUsers(rr2, req2)

	assert.Equal(t, http.StatusOK, rr2.Code)
	var users []User
	json.NewDecoder(rr2.Body).Decode(&users)
	require.Len(t, users, 1)
	assert.Equal(t, 150.50, users[0].Balance)
}
