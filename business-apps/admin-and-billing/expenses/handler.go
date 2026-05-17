package expenses

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/model"
)

func GetExpenses(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	query := `SELECT EXPENSE_ID, EXPENSE_DATE, REASON, AMOUNT, CREATED_AT FROM EXPENSES`
	var args []interface{}

	if startDateStr != "" && endDateStr != "" {
		startDate, _ := time.Parse("2006-01-02", startDateStr)
		endDate, _ := time.Parse("2006-01-02", endDateStr)
		query += ` WHERE EXPENSE_DATE BETWEEN $1 AND $2`
		args = append(args, startDate, endDate)
	}

	query += ` ORDER BY EXPENSE_DATE DESC, CREATED_AT DESC`

	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var expenses []model.Expense
	for rows.Next() {
		var e model.Expense
		err := rows.Scan(&e.ExpenseID, &e.ExpenseDate, &e.Reason, &e.Amount, &e.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		expenses = append(expenses, e)
	}

	json.NewEncoder(w).Encode(expenses)
}

func CreateExpense(w http.ResponseWriter, r *http.Request) {
	var e model.Expense
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbPool := database.GetDbConn()
	err := dbPool.QueryRow(r.Context(), `
		INSERT INTO EXPENSES (EXPENSE_DATE, REASON, AMOUNT)
		VALUES ($1, $2, $3)
		RETURNING EXPENSE_ID, CREATED_AT
	`, e.ExpenseDate, e.Reason, e.Amount).Scan(&e.ExpenseID, &e.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(e)
}

func UpdateExpense(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	var e model.Expense
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbPool := database.GetDbConn()
	_, err := dbPool.Exec(r.Context(), `
		UPDATE EXPENSES SET EXPENSE_DATE = $1, REASON = $2, AMOUNT = $3
		WHERE EXPENSE_ID = $4
	`, e.ExpenseDate, e.Reason, e.Amount, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteExpense(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	dbPool := database.GetDbConn()
	_, err := dbPool.Exec(r.Context(), `DELETE FROM EXPENSES WHERE EXPENSE_ID = $1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
