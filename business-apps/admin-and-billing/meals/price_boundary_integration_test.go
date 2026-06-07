package meals

import (
	"context"
	"testing"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

// TestPriceBoundary_LocalMidnight verifies that a price change scheduled at
// local midnight (Asia/Kolkata) takes effect at the correct instant (converted to UTC)
// and that lookups just before/after that instant return the expected prices.
func TestPriceBoundary_LocalMidnight(t *testing.T) {
	testdb.ResetData()

	// choose a fixed date in the future to avoid colliding with seeded backfill
	loc, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)

	// effective local midnight: 2026-05-15 00:00 IST
	effLocal := time.Date(2026, 5, 15, 0, 0, 0, 0, loc)
	// convert to UTC instant that will be stored in DB
	effUTC := effLocal.UTC()

	ctx := context.Background()
	// read old price as of just before the effective instant
	oldPrice, err := GetPriceAtForItem(ctx, "standard", effUTC.Add(-1*time.Second))
	require.NoError(t, err)

	newPrice := oldPrice + 10.0
	require.NoError(t, InsertPriceHistory(ctx, "standard", newPrice, effUTC, "test"))

	// lookup just before effective instant -> should get oldPrice
	pBefore, err := GetPriceAtForItem(ctx, "standard", effUTC.Add(-1*time.Second))
	require.NoError(t, err)
	require.Equal(t, oldPrice, pBefore)

	// lookup exactly at effective instant -> should get newPrice (effective_from <= ts)
	pAt, err := GetPriceAtForItem(ctx, "standard", effUTC)
	require.NoError(t, err)
	require.Equal(t, newPrice, pAt)

	// lookup after effective instant -> newPrice
	pAfter, err := GetPriceAtForItem(ctx, "standard", effUTC.Add(1*time.Second))
	require.NoError(t, err)
	require.Equal(t, newPrice, pAfter)

	// Also verify GetPricesAt returns the new price when queried after effective instant
	pricesMap, err := GetPricesAt(ctx, effUTC.Add(1*time.Second))
	require.NoError(t, err)
	reqPrice, ok := pricesMap["standard"]
	require.True(t, ok)
	require.Equal(t, newPrice, reqPrice)
}
