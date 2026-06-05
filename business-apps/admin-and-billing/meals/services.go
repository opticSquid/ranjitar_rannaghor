package meals

import (
	"context"
	"log/slog"
	"time"
)

// Maintained signature for journal module internal use
func GetMealPricesInternal(ctx context.Context) map[string]float64 {
	prices, err := FetchMealPricesInternal(ctx)
	if err != nil {
		slog.Error("Failed to get meal prices", "err", err)
	}
	return prices
}

// GetMealPricesAt returns prices effective at the provided timestamp.
func GetMealPricesAt(ctx context.Context, ts time.Time) map[string]float64 {
	prices, err := GetPricesAt(ctx, ts.UTC())
	if err != nil {
		slog.Error("Failed to get meal prices at time", "err", err, "ts", ts)
		// Fallback to current prices
		p, _ := FetchMealPricesInternal(ctx)
		return p
	}
	return prices
}

func CreateMealService(ctx context.Context, m *MealPrice) error {
	return InsertMeal(ctx, m)
}

func GetMealsService(ctx context.Context) ([]MealPrice, error) {
	return FetchMeals(ctx)
}

func UpdateMealService(ctx context.Context, id string, price float64) error {
	return UpdateMealPriceInDB(ctx, id, price)
}

func DeleteMealService(ctx context.Context, id string) error {
	return DeleteMealFromDB(ctx, id)
}
