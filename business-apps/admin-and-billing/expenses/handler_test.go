package expenses

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
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/model"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testdb.Setup()
	defer testdb.Teardown()
	m.Run()
}

func TestCreateExpense_Success(t *testing.T) {
	testdb.ResetData()

	reqBody := model.Expense{
		ExpenseDate: time.Now().Truncate(24 * time.Hour),
		Reason:      "Gas Cylinder",
		Amount:      1050.0,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/expenses", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	CreateExpense(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp model.Expense
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.True(t, resp.ExpenseID > 0)
	assert.Equal(t, "Gas Cylinder", resp.Reason)
	assert.Equal(t, 1050.0, resp.Amount)
}

func TestGetExpenses_All(t *testing.T) {
	testdb.ResetData()

	_, err := testdb.DbPool.Exec(context.Background(), `
		INSERT INTO expenses (expense_date, reason, amount) VALUES
		('2023-10-01', 'Vegetables', 200.0),
		('2023-10-02', 'Chicken', 500.0)
	`)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/expenses", nil)
	rr := httptest.NewRecorder()

	GetExpenses(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var expenses []model.Expense
	err = json.NewDecoder(rr.Body).Decode(&expenses)
	require.NoError(t, err)

	assert.Len(t, expenses, 2)
}

func TestGetExpenses_DateRange(t *testing.T) {
	testdb.ResetData()

	_, err := testdb.DbPool.Exec(context.Background(), `
		INSERT INTO expenses (expense_date, reason, amount) VALUES
		('2023-09-30', 'Out of range', 100.0),
		('2023-10-05', 'In range', 200.0),
		('2023-10-15', 'Out of range later', 300.0)
	`)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/expenses?start_date=2023-10-01&end_date=2023-10-10", nil)
	rr := httptest.NewRecorder()

	GetExpenses(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var expenses []model.Expense
	json.NewDecoder(rr.Body).Decode(&expenses)

	require.Len(t, expenses, 1)
	assert.Equal(t, "In range", expenses[0].Reason)
}

func TestUpdateExpense_Success(t *testing.T) {
	testdb.ResetData()

	var id int
	err := testdb.DbPool.QueryRow(context.Background(), `
		INSERT INTO expenses (expense_date, reason, amount) VALUES ('2023-10-01', 'Old Reason', 100.0) RETURNING expense_id
	`).Scan(&id)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Put("/expenses/{id}", UpdateExpense)

	reqBody := model.Expense{
		ExpenseDate: time.Now().Truncate(24 * time.Hour),
		Reason:      "New Reason",
		Amount:      150.0,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/expenses/%d", id), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var updatedReason string
	var updatedAmount float64
	testdb.DbPool.QueryRow(context.Background(), `SELECT reason, amount FROM expenses WHERE expense_id = $1`, id).Scan(&updatedReason, &updatedAmount)

	assert.Equal(t, "New Reason", updatedReason)
	assert.Equal(t, 150.0, updatedAmount)
}

func TestDeleteExpense_Success(t *testing.T) {
	testdb.ResetData()

	var id int
	err := testdb.DbPool.QueryRow(context.Background(), `
		INSERT INTO expenses (expense_date, reason, amount) VALUES ('2023-10-01', 'To delete', 100.0) RETURNING expense_id
	`).Scan(&id)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Delete("/expenses/{id}", DeleteExpense)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/expenses/%d", id), nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var count int
	testdb.DbPool.QueryRow(context.Background(), `SELECT count(*) FROM expenses WHERE expense_id = $1`, id).Scan(&count)
	assert.Equal(t, 0, count)
}
