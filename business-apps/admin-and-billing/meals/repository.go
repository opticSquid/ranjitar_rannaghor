package meals

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func (c MenuItemCategory) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

func (c *MenuItemCategory) UnmarshalText(text []byte) error {
	val := MenuItemCategory(text)
	switch val {
	case COMBO_THALI, A_LA_CARTE:
		*c = val
		return nil
	default:
		return fmt.Errorf("invalid MenuItemCategory value: %s", string(text))
	}
}

func FetchMealPricesInternal(ctx context.Context, date time.Time, menu_items []string) (map[int]float64, error) {
	prices := make(map[int]float64)
	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(ctx, "SELECT DISTINCE ON (ITEM_ID) ITEM_ID, PRICE FROM MEAL_PRICE_HISTORY WHERE ITEM_NAME = ANY($1) and EFFECTIVE_FROM <= $2 ORDER BY ITEM_ID, EFFECTIVE_FROM DESC", menu_items, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var price float64
		if err := rows.Scan(&id, &price); err != nil {
			return nil, err
		}
		prices[id] = price
	}
	return prices, nil
}

func InsertMenuItem(tx pgx.Tx, ctx context.Context, m *MenuItem) error {
	return tx.QueryRow(ctx, `
			INSERT INTO MENU_ITEMS (ITEM_NAME, CATEGORY, IS_ACTIVE)
			VALUES ($1, $2, $3)
			RETURNING ITEM_ID
		`, m.name, m.category, m.isActive).Scan(m.itemId)
}

func InsertMenuItemPrice(tx pgx.Tx, ctx context.Context, p *MenuPriceHistory) error {
	return tx.QueryRow(ctx, `
		INSERT INTO MENU_PRICE_SCHEDULE (ITEM_ID, PRICE, EFFECTIVE_FROM)
		VALUES ($1, $2, $3)
		RETURNING PRICE_ID
		`, p.itemId, p.price, p.effectiveFrom).Scan(p.priceId)
}

func FetchMeals(ctx context.Context) ([]MenuItem, error) {
	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(ctx, "SELECT ITEM_ID, ITEM_NAME, PRICE, UPDATED_AT FROM MEAL_PRICES ORDER BY PRICE DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []MenuItem
	for rows.Next() {
		var p MenuItem
		err := rows.Scan(&p.ItemID, &p.name, &p.Price, &p.UpdatedAt)
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
func FetchPriceHistoryForItem(ctx context.Context, itemID string) ([]MenuPriceHistory, error) {
	var entries []MenuPriceHistory
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
		var e MenuPriceHistory
		if err := rows.Scan(&e.ID, &e.itemId, &e.price, &e.effectiveFrom, &e.CreatedBy, &e.createdAt); err != nil {
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
