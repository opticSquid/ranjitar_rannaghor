package expenses

import (
	"context"
)

func GetExpensesService(ctx context.Context, startDateStr, endDateStr string) ([]Expense, error) {
	return FetchExpensesFromDB(ctx, startDateStr, endDateStr)
}

func CreateExpenseService(ctx context.Context, e *Expense) error {
	return InsertExpenseToDB(ctx, e)
}

func UpdateExpenseService(ctx context.Context, id int, e *Expense) error {
	return UpdateExpenseInDB(ctx, id, e)
}

func DeleteExpenseService(ctx context.Context, id int) error {
	return DeleteExpenseFromDB(ctx, id)
}
