package expenses

import (
	"context"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func FetchExpensesFromDB(ctx context.Context, startDateStr, endDateStr string) ([]Expense, error) {
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
	rows, err := dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var e Expense
		err := rows.Scan(&e.ExpenseID, &e.ExpenseDate, &e.Reason, &e.Amount, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

func InsertExpenseToDB(ctx context.Context, e *Expense) error {
	dbPool := database.GetDbConn()
	return dbPool.QueryRow(ctx, `
		INSERT INTO EXPENSES (EXPENSE_DATE, REASON, AMOUNT)
		VALUES ($1, $2, $3)
		RETURNING EXPENSE_ID, CREATED_AT
	`, e.ExpenseDate, e.Reason, e.Amount).Scan(&e.ExpenseID, &e.CreatedAt)
}

func UpdateExpenseInDB(ctx context.Context, id int, e *Expense) error {
	dbPool := database.GetDbConn()
	_, err := dbPool.Exec(ctx, `
		UPDATE EXPENSES SET EXPENSE_DATE = $1, REASON = $2, AMOUNT = $3
		WHERE EXPENSE_ID = $4
	`, e.ExpenseDate, e.Reason, e.Amount, id)
	return err
}

func DeleteExpenseFromDB(ctx context.Context, id int) error {
	dbPool := database.GetDbConn()
	_, err := dbPool.Exec(ctx, `DELETE FROM EXPENSES WHERE EXPENSE_ID = $1`, id)
	return err
}
