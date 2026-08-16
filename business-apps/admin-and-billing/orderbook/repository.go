package orderbook

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/utils"
)

func (c MealType) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

func (c *MealType) UnmarshalText(text []byte) error {
	val := MealType(text)
	switch val {
	case LUNCH, DINNER:
		*c = val
		return nil
	default:
		return fmt.Errorf("invalid MealType value: %s", string(text))
	}
}

func checkUserExist(tx pgx.Tx, ctx context.Context, userId int) (bool, error) {
	var doesExist bool
	err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM PUBLIC.USERS WHERE USER_ID = $1)", userId).Scan(&doesExist)
	if err != nil {
		return false, err
	}
	return doesExist, nil
}

func checkMenuItemsExist(tx pgx.Tx, ctx context.Context, menuItemIds []int) ([]int, error) {
	rows, err := tx.Query(ctx, `SELECT
		INPUT_ID
	FROM
		UNNEST($1::INT[]) AS INPUT_ID
	WHERE
		NOT EXISTS (
			SELECT
				1
			FROM
				PUBLIC.MENU_ITEMS
			WHERE
				ITEM_ID = INPUT_ID
		)`, menuItemIds)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var invalidItemIds []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan missing item ID: %w", err)
		}
		invalidItemIds = append(invalidItemIds, id)
	}
	// Check for any error encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating validation rows: %w", err)
	}
	return invalidItemIds, nil
}

func fetchPrices(tx pgx.Tx, ctx context.Context, itemIds []int, ts time.Time) (map[int]price, error) {
	rows, err := tx.Query(ctx, `SELECT DISTINCT
		ON (ITEM_ID) ITEM_ID,
		PRICE_ID,
		PRICE
	FROM
		PUBLIC.MENU_PRICE_SCHEDULE
	WHERE
		ITEM_ID = ANY ($1)
		AND EFFECTIVE_FROM <= $2
	ORDER BY
		ITEM_ID,
		EFFECTIVE_FROM DESC;`, itemIds, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices %w", err)
	}
	defer rows.Close()

	prices := make(map[int]price, len(itemIds))
	for rows.Next() {
		var itemId int
		var price price
		if err := rows.Scan(&itemId, &price.priceId, &price.price); err != nil {
			return nil, fmt.Errorf("failed to scan price %w", err)
		}
		prices[itemId] = price
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to fetch prices %w", rows.Err())
	}
	if len(prices) != len(itemIds) {
		return nil, fmt.Errorf("expected %d prices, got %d", len(itemIds), len(prices))
	}
	return prices, nil
}

func createInitialOrder(tx pgx.Tx, ctx context.Context, order order) (int, error) {
	var orderId int
	err := tx.QueryRow(ctx, `INSERT INTO
		PUBLIC.ORDERS (
			USER_ID,
			ORDER_TIMESTAMP,
			MEAL_TYPE,
			TOTAL_AMOUNT,
			STATUS
		)
	VALUES
		($1, $2, $3, $4, $5)
	RETURNING
		ORDER_ID`, order.userId, order.orderTs, order.mealType, order.totalAmount, order.status).Scan(&orderId)
	if err != nil {
		return 0, err
	}
	return orderId, nil
}

func createOrderDetails(tx pgx.Tx, ctx context.Context, orderItems []orderItem) error {
	rows := [][]interface{}{}
	for _, item := range orderItems {
		rows = append(rows, []interface{}{item.orderId, item.itemId, item.priceId, item.quantity, item.subTotal})
	}
	// bulk insert using copy from operation
	copyCount, err := tx.CopyFrom(ctx,
		pgx.Identifier{"PUBLIC.ORDER_ITEMS"},
		[]string{"ORDER_ID", "ITEM_ID", "PRICE_ID", "QUANTITY", "SUBTOTAL"},
		pgx.CopyFromRows(rows))
	if err != nil {
		return err
	}
	if copyCount != int64(len(rows)) {
		return fmt.Errorf("expected to copy %d rows, got %d", len(rows), copyCount)
	}
	return nil
}

func createTransaction(tx pgx.Tx, ctx context.Context, txn walletTransaction) error {
	res, err := tx.Exec(ctx, `INSERT INTO
		PUBLIC.WALLET_TRANSACTIONS (
			USER_ID,
			TXN_TYPE,
			AMOUNT,
			TXN_TIMESTAMP,
			ORDER_ID
		)
	VALUES
		($1, $2, $3, $4, $5)
	`, txn.userId, txn.txnType, txn.amount, txn.txnTs, txn.orderId)
	if err != nil {
		return err
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("expected to insert 1 row, got %d", res.RowsAffected())
	}
	return nil
}

func createFinalizedOrder(tx pgx.Tx, ctx context.Context, order order) error {
	res, err := tx.Exec(ctx, `UPDATE PUBLIC.ORDERS SET TOTAL_AMOUNT = $1, STATUS = $2 WHERE ORDER_ID = $3`, order.totalAmount, order.status, order.orderId)
	if err != nil {
		return err
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("expected to update 1 row, got %d", res.RowsAffected())
	}
	return nil
}

func DeleteDailyEntryFromDB(ctx context.Context, logID int) (float64, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var userID int
	var totalCost float64
	var logDate time.Time
	err = tx.QueryRow(ctx, `SELECT USER_ID, TOTAL_COST, LOG_DATE FROM DAILY_LOGS WHERE LOG_ID = $1`, logID).Scan(&userID, &totalCost, &logDate)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, errors.New("Entry not found")
		}
		return 0, err
	}

	_, err = tx.Exec(ctx, `DELETE FROM DAILY_LOGS WHERE LOG_ID = $1`, logID)
	if err != nil {
		return 0, err
	}

	createdAt := getCreationTime(logDate)

	var prevBalanceAfter *float64
	err = tx.QueryRow(ctx, `SELECT BALANCE_AFTER FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 AND CREATED_AT < $2 ORDER BY CREATED_AT DESC LIMIT 1`, userID, createdAt).Scan(&prevBalanceAfter)

	if err != nil && err.Error() != "no rows in result set" {
		return 0, err
	}

	var currentBalance float64 = 0
	if prevBalanceAfter != nil {
		currentBalance = *prevBalanceAfter
	}
	newBalance := currentBalance + totalCost

	_, err = tx.Exec(ctx, `
		INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, CREATED_AT)
		VALUES ($1, $2, 'confirmed', $3, $4, $5)
	`, userID, utils.REFUND, totalCost, newBalance, createdAt)
	if err != nil {
		return 0, err
	}

	err = utils.RecalculateBalances(ctx, tx, utils.REFUND, userID, createdAt, totalCost)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return newBalance, nil
}

func FetchDailyEntries(ctx context.Context, date time.Time, userID int) ([]DailyLog, error) {
	query := `
		SELECT l.LOG_ID, l.USER_ID, u.NAME as USER_NAME, l.LOG_DATE, l.MEAL_TYPE,
		       l.HAS_MAIN_MEAL, l.IS_SPECIAL, l.SPECIAL_DISH_NAME,
		       l.EXTRA_RICE_QTY, l.EXTRA_ROTI_QTY, l.EXTRA_CHICKEN_QTY, l.EXTRA_FISH_QTY, l.EXTRA_EGG_QTY, l.EXTRA_VEGETABLE_QTY, l.TOTAL_COST
		FROM DAILY_LOGS l
		JOIN USERS u ON l.USER_ID = u.USER_ID
		WHERE l.LOG_DATE = $1
	`
	args := []any{date}

	if userID != 0 {
		query += " AND l.USER_ID = $2"
		args = append(args, userID)
	}

	query += " ORDER BY u.NAME ASC, l.MEAL_TYPE DESC"

	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []DailyLog
	for rows.Next() {
		var l DailyLog
		err := rows.Scan(&l.LogID, &l.UserID, &l.UserName, &l.LogDate, &l.MealType, &l.HasMainMeal, &l.IsSpecial, &l.SpecialDishName, &l.ExtraRiceQty, &l.ExtraRotiQty, &l.ExtraChickenQty, &l.ExtraFishQty, &l.ExtraEggQty, &l.ExtraVegetableQty, &l.TotalCost)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func fetchActiveMenuItems(tx pgx.Tx, ctx context.Context) ([]MenuItemResponse, error) {
	rows, err := tx.Query(ctx, `SELECT DISTINCT
		ON (I.ITEM_ID) I.ITEM_ID,
		I.ITEM_NAME,
		I.CATEGORY,
		P.PRICE,
		P.EFFECTIVE_FROM
	FROM
		PUBLIC.MENU_ITEMS I
		JOIN PUBLIC.MENU_PRICE_SCHEDULE P ON I.ITEM_ID = P.ITEM_ID
	WHERE
		I.IS_ACTIVE = TRUE
		AND P.EFFECTIVE_FROM <= NOW()
	ORDER BY
		I.ITEM_ID,
		P.EFFECTIVE_FROM DESC;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MenuItemResponse
	for rows.Next() {
		var item MenuItemResponse
		err := rows.Scan(&item.ItemId, &item.ItemName, &item.Category, &item.LatestPrice, &item.EffectiveFrom)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
