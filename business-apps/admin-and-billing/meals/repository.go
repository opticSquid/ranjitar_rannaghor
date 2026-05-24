package meals

import (
	"context"

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
