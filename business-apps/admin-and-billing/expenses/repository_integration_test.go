package expenses

import (
	"context"
	"testing"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestInsertAndFetchExpense(t *testing.T) {
	testdb.ResetData()
	e := &Expense{ExpenseDate: time.Now().Truncate(24 * time.Hour), Reason: "Fuel", Amount: 250.0}
	err := InsertExpenseToDB(context.Background(), e)
	require.NoError(t, err)
	require.True(t, e.ExpenseID > 0)

	exps, err := FetchExpensesFromDB(context.Background(), "", "")
	require.NoError(t, err)
	require.Len(t, exps, 1)
}
