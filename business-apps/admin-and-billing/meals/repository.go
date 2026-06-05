package meals

import (
	"context"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func FetchMealPricesInternal(ctx context.Context) (map[string]float64, error) {
	prices := make(map[string]float64)
	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(ctx, "SELECT ITEM_ID, PRICE FROM MEAL_PRICES")
	if err != nil {
		return prices, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var price float64
		if err := rows.Scan(&id, &price); err == nil {
			prices[id] = price
		}
	}
	return prices, nil
}

func InsertMeal(ctx context.Context, m *MealPrice) error {
	dbPool := database.GetDbConn()
	if m.ItemID == "" {
		// generate a UUID for item id
		id := utilsGenerateUUID()
		m.ItemID = id
	}
	return dbPool.QueryRow(ctx, `
			INSERT INTO MEAL_PRICES (ITEM_ID, ITEM_NAME, PRICE)
			VALUES ($1, $2, $3)
			RETURNING ITEM_ID
		`, m.ItemID, m.ItemName, m.Price).Scan(&m.ItemID)
}

func FetchMeals(ctx context.Context) ([]MealPrice, error) {
	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(ctx, "SELECT ITEM_ID, ITEM_NAME, PRICE, UPDATED_AT FROM MEAL_PRICES ORDER BY PRICE DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []MealPrice
	for rows.Next() {
		var p MealPrice
		err := rows.Scan(&p.ItemID, &p.ItemName, &p.Price, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, nil
}

func UpdateMealPriceInDB(ctx context.Context, id string, price float64) error {
	dbPool := database.GetDbConn()
	_, err := dbPool.Exec(ctx, `
		UPDATE MEAL_PRICES SET PRICE = $1, UPDATED_AT = CURRENT_TIMESTAMP
		WHERE ITEM_ID = $2
	`, price, id)
	return err
}

func DeleteMealFromDB(ctx context.Context, id string) error {
	dbPool := database.GetDbConn()
	_, err := dbPool.Exec(ctx, `
		DELETE FROM MEAL_PRICES
		WHERE ITEM_ID = $1
	`, id)
	return err
}

// InsertPriceHistory inserts a historical price change for an item.
func InsertPriceHistory(ctx context.Context, itemID string, price float64, effectiveFrom time.Time, createdBy string) error {
	dbPool := database.GetDbConn()
	_, err := dbPool.Exec(ctx, `
		INSERT INTO meal_price_history (item_id, price, effective_from, created_by)
		VALUES ($1, $2, $3, $4)
	`, itemID, price, effectiveFrom, createdBy)
	return err
}

// FetchPriceHistoryForItem returns all history entries for an item ordered by effective_from desc.
func FetchPriceHistoryForItem(ctx context.Context, itemID string) ([]PriceHistoryEntry, error) {
	var entries []PriceHistoryEntry
	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(ctx, `
		SELECT id, item_id, price, effective_from, created_by, created_at
		FROM meal_price_history
		WHERE item_id = $1
		ORDER BY effective_from DESC
	`, itemID)
	if err != nil {
		return entries, err
	}
	defer rows.Close()

	for rows.Next() {
		var e PriceHistoryEntry
		if err := rows.Scan(&e.ID, &e.ItemID, &e.Price, &e.EffectiveFrom, &e.CreatedBy, &e.CreatedAt); err != nil {
			return entries, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// GetPriceAtForItem returns the price for an item effective at a given timestamp.
func GetPriceAtForItem(ctx context.Context, itemID string, ts time.Time) (float64, error) {
	dbPool := database.GetDbConn()
	var price float64
	err := dbPool.QueryRow(ctx, `
		SELECT price FROM meal_price_history
		WHERE item_id = $1 AND effective_from <= $2
		ORDER BY effective_from DESC LIMIT 1
	`, itemID, ts).Scan(&price)
	return price, err
}

// GetPricesAt returns a map of item_id -> price effective at timestamp ts.
func GetPricesAt(ctx context.Context, ts time.Time) (map[string]float64, error) {
	prices := make(map[string]float64)
	dbPool := database.GetDbConn()
	// debug: log the timestamp used for lookup
	// fmt.Printf("GetPricesAt: looking up prices at %v\n", ts)
	rows, err := dbPool.Query(ctx, `
		SELECT item_id, price FROM (
			SELECT item_id, price,
			ROW_NUMBER() OVER (PARTITION BY item_id ORDER BY effective_from DESC) AS rn
			FROM meal_price_history
			WHERE effective_from <= $1
		) q WHERE rn = 1
	`, ts)
	if err != nil {
		return prices, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var price float64
		if err := rows.Scan(&id, &price); err == nil {
			prices[id] = price
		}
	}
	return prices, nil
}
