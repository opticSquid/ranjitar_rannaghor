package meals

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

// Maintained signature for journal module internal use
func GetMealPricesInternal(ctx context.Context, date time.Time, menu_items []string) map[string]float64 {
	prices, err := FetchMealPricesInternal(ctx, date, menu_items)
	if err != nil {
		slog.Error("Failed to get meal prices", "err", err)
	}
	return prices
}

// GetMealPricesAt returns prices effective at the provided timestamp.
func GetMealPricesAt(ctx context.Context, ts time.Time) map[string]float64 {
	prices, err := GetPricesAt(ctx, ts)
	if err != nil {
		slog.Error("Failed to get meal prices at time", "err", err, "ts", ts)
		// Fallback to current prices
		p, _ := FetchMealPricesInternal(ctx, ts, make([]string, 0))
		return p
	}
	return prices
}

func CreateMenuItemService(ctx context.Context, r *NewMenuItemRequest) error {
	item := &MenuItem{
		name:     r.Name,
		isActive: true,
	}
	itemPriceEntry := &MenuPriceHistory{
		price: r.Price,
	}

	// Data validation
	switch r.Category {
	case "combo_thali":
		item.category = COMBO_THALI
	case "a_la_carte":
		item.category = A_LA_CARTE
	default:
		return fmt.Errorf("menu item category is invalid. Category: %s.  %w", r.Category, ErrInvalidMenuCategory)
	}
	curTime := time.Now().UTC()
	if r.EffectiveFrom.UTC().Before(curTime) {
		return fmt.Errorf("effective_from value can not be in the past of current time. current timestamp (utc): %v, effective_from value (utc): %v; %w", curTime, r.EffectiveFrom.UTC(), ErrEffectiveFromValueOfPast)
	}
	itemPriceEntry.effectiveFrom = r.EffectiveFrom

	// persisting data
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	//check if an item with the same name already exists in menu then fail
	does_exist := false
	err = CheckMenuItemExistance(tx, ctx, item, &does_exist)
	if err != nil {
		return fmt.Errorf("failed to check menu item existence: %w", err)
	}
	if does_exist {
		return fmt.Errorf("menu item with name %s already exists", item.name, ErrMenuItemExists)
	}
	// relating data
	// item.itemId will be populated if above transaction succeeds
	itemPriceEntry.itemId = item.itemId
	err = InsertMenuItem(tx, ctx, item)
	if err != nil {
		return fmt.Errorf("failed to create menu item: %w", err)
	}

	err = InsertMenuItemPrice(tx, ctx, itemPriceEntry)
	if err != nil {
		return fmt.Errorf("failed to insert menu item price: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func GetMealsService(ctx context.Context) ([]MenuItem, error) {
	return FetchMeals(ctx)
}

func UpdateMealService(ctx context.Context, id string, price float64) error {
	return UpdateMealPriceInDB(ctx, id, price)
}

func DeleteMealService(ctx context.Context, id string) error {
	return DeleteMealFromDB(ctx, id)
}
