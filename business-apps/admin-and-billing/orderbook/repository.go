package orderbook

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/meals"
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

func InsertWalletTxn(ctx context.Context, txns []walletTxn) (int64, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	table := pgx.Identifier{"wallet_transactions"}
	columns := []string{"user_id", "txn_date", "txn_type", "meal_type", "dish_name", "quantity", "amount", "menu_item_id"}

	rows := [][]interface{}{}
	for _, txn := range txns {
		rows = append(rows, []interface{}{txn.userId, txn.txnDate, txn.txnType, txn.mealType, txn.dishName, txn.quantity, txn.amount, txn.menuItemId})
	}

	rowsAffected, err := tx.CopyFrom(ctx, table, columns, pgx.CopyFromRows(rows))
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
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

func UpdateDailyEntryInDB(ctx context.Context, logID int, req EntryRequest) (float64, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var userID int
	var oldTotalCost float64
	var logDate time.Time
	var logCreationTime time.Time
	err = tx.QueryRow(ctx, `SELECT USER_ID, TOTAL_COST, LOG_DATE, CREATED_AT FROM DAILY_LOGS WHERE LOG_ID = $1`, logID).Scan(&userID, &oldTotalCost, &logDate, &logCreationTime)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, errors.New("Entry not found")
		}
		return 0, err
	}

	// compute new total cost using price history at the original creation timestamp
	createdAtForPrice := constructCreationTime(logDate, logCreationTime).UTC()
	prices := meals.GetMealPricesAt(ctx, createdAtForPrice)
	newTotalCost := CalculateTotalCost(req, prices)

	if newTotalCost != oldTotalCost {
		// construct createdAt to have date from logDate and time from logCreationTime
		createdAt := constructCreationTime(logDate, logCreationTime)

		var prevBalanceAfter *float64
		var maxCreatedAt time.Time = createdAt
		err = tx.QueryRow(ctx, `
			SELECT BALANCE_AFTER, CREATED_AT
			FROM WALLET_TRANSACTIONS
			WHERE USER_ID = $1 AND CREATED_AT >= $2 AND CREATED_AT < $3
			ORDER BY CREATED_AT DESC, TXN_ID DESC LIMIT 1
		`, userID, createdAt, createdAt.Add(1*time.Minute)).Scan(&prevBalanceAfter, &maxCreatedAt)
		if err != nil && err.Error() == "no rows in result set" {
			return 0, err
		}
		var prevBalance float64 = 0
		if prevBalanceAfter != nil {
			prevBalance = *prevBalanceAfter
		}

		txBalanceAfter := prevBalance + oldTotalCost
		// ensure refund is sent strictly after the latest adjustment
		createdAt = maxCreatedAt.Add(1 * time.Microsecond)
		_, err = tx.Exec(ctx, `
			INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, CREATED_AT)
			VALUES ($1, $2, 'confirmed', $3, $4, $5)
		`, userID, utils.REFUND, oldTotalCost, txBalanceAfter, createdAt)

		if err != nil {
			return 0, err
		}
		err = utils.RecalculateBalances(ctx, tx, utils.REFUND, userID, createdAt, oldTotalCost)
		if err != nil {
			return 0, err
		}

		txBalanceAfter -= newTotalCost
		// ensure new delivery is sent after refund record
		createdAt = createdAt.Add(1 * time.Microsecond)
		_, err = tx.Exec(ctx, `
			INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, CREATED_AT)
			VALUES ($1, $2, 'confirmed', $3, $4, $5)
		`, userID, utils.DELIVERY, newTotalCost, txBalanceAfter, createdAt)

		if err != nil {
			return 0, err
		}
		err = utils.RecalculateBalances(ctx, tx, utils.DELIVERY, userID, createdAt, newTotalCost)
		if err != nil {
			return 0, err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE DAILY_LOGS
		SET MEAL_TYPE = $1, HAS_MAIN_MEAL = $2, IS_SPECIAL = $3, SPECIAL_DISH_NAME = $4, EXTRA_RICE_QTY = $5, EXTRA_ROTI_QTY = $6, TOTAL_COST = $7
		WHERE LOG_ID = $8
	`, req.MealType, req.HasMainMeal, req.IsSpecial, req.IsSpecialMenu, req.ExtraRiceQty, req.ExtraRotiQty, newTotalCost, logID)
	if err != nil {
		return 0, err
	}

	var finalBalance float64
	err = tx.QueryRow(ctx, `SELECT COALESCE(BALANCE_AFTER, 0) FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 ORDER BY CREATED_AT DESC, TXN_ID DESC LIMIT 1`, userID).Scan(&finalBalance)
	if err != nil {
		if err.Error() == "no rows in result set" {
			finalBalance = 0
		} else {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return finalBalance, nil
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
