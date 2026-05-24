package meals

import (
	"context"
	"testing"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestInsertAndFetchMeal(t *testing.T) {
	testdb.ResetData()
	m := &MealPrice{ItemName: "RepoMeal", Price: 3.5}
	err := InsertMeal(context.Background(), m)
	require.NoError(t, err)
	require.NotEmpty(t, m.ItemID)

	meals, err := FetchMeals(context.Background())
	require.NoError(t, err)
	require.True(t, len(meals) >= 1)
}
