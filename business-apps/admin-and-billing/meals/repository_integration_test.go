package meals

import (
	"context"
	"testing"
	"time"

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

func TestPriceHistoryEffectiveTime(t *testing.T) {
	testdb.ResetData()
	ctx := context.Background()
	now := time.Now().UTC()

	// fetch current price
	oldPrice, err := GetPriceAtForItem(ctx, "standard", now)
	require.NoError(t, err)

	newPrice := oldPrice + 10.0
	eff := now.Add(1 * time.Minute)
	require.NoError(t, InsertPriceHistory(ctx, "standard", newPrice, eff, "test"))

	// price before effective time should be old price
	p1, err := GetPriceAtForItem(ctx, "standard", now)
	require.NoError(t, err)
	require.Equal(t, oldPrice, p1)

	// price at or after effective time should be new price
	p2, err := GetPriceAtForItem(ctx, "standard", eff)
	require.NoError(t, err)
	require.Equal(t, newPrice, p2)

	p3Map, err := GetPricesAt(ctx, eff.Add(2*time.Minute))
	require.NoError(t, err)
	require.Equal(t, newPrice, p3Map["standard"])
}
