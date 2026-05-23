package journal

import (
	"context"
	"errors"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/utils"
)

func CreateDailyEntryInDB(ctx context.Context, log EntryRequest, totalCost float64, createdAt time.Time) (float64, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Insert Log
	_, err = tx.Exec(ctx, `
		INSERT INTO DAILY_LOGS (USER_ID, LOG_DATE, MEAL_TYPE, HAS_MAIN_MEAL, IS_SPECIAL, SPECIAL_DISH_NAME, EXTRA_RICE_QTY, EXTRA_ROTI_QTY, EXTRA_CHICKEN_QTY, EXTRA_FISH_QTY, EXTRA_EGG_QTY, EXTRA_VEGETABLE_QTY, TOTAL_COST)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, log.UserID, log.LogDate, log.MealType, log.HasMainMeal, log.IsSpecial, log.SpecialDishName, log.ExtraRiceQty, log.ExtraRotiQty, log.ExtraChickenQty, log.ExtraFishQty, log.ExtraEggQty, log.ExtraVegetableQty, totalCost)
	if err != nil {
		return 0, err
	}

	var prevBalanceAfter *float64
	err = tx.QueryRow(ctx, `SELECT BALANCE_AFTER FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 AND CREATED_AT < $2 ORDER BY CREATED_AT DESC LIMIT 1`, log.UserID, createdAt).Scan(&prevBalanceAfter)

	if err != nil && err.Error() != "no rows in result set" {
		return 0, err
	}

	var currentBalance float64 = 0
	if prevBalanceAfter != nil {
		currentBalance = *prevBalanceAfter
	}
	newBalance := currentBalance - totalCost

	_, err = tx.Exec(ctx, `
		INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, CREATED_AT)
		VALUES ($1, 'delivery', 'confirmed', $2, $3, $4)
	`, log.UserID, totalCost, newBalance, createdAt)
	if err != nil {
		return 0, err
	}

	err = utils.RecalculateBalances(ctx, tx, utils.DELIVERY, log.UserID, createdAt, totalCost)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return newBalance, nil
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
		VALUES ($1, 'refund', 'confirmed', $2, $3, $4)
	`, userID, totalCost, newBalance, createdAt)
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

func UpdateDailyEntryInDB(ctx context.Context, logID int, req EntryRequest, newTotalCost float64) (float64, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var userID int
	var oldTotalCost float64
	var logDate time.Time
	err = tx.QueryRow(ctx, `SELECT USER_ID, TOTAL_COST, LOG_DATE FROM DAILY_LOGS WHERE LOG_ID = $1`, logID).Scan(&userID, &oldTotalCost, &logDate)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, errors.New("Entry not found")
		}
		return 0, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE DAILY_LOGS
		SET MEAL_TYPE = $1, HAS_MAIN_MEAL = $2, IS_SPECIAL = $3, SPECIAL_DISH_NAME = $4, EXTRA_RICE_QTY = $5, EXTRA_ROTI_QTY = $6, TOTAL_COST = $7
		WHERE LOG_ID = $8
	`, req.MealType, req.HasMainMeal, req.IsSpecial, req.SpecialDishName, req.ExtraRiceQty, req.ExtraRotiQty, newTotalCost, logID)
	if err != nil {
		return 0, err
	}

	if newTotalCost != oldTotalCost {
		createdAt := getCreationTime(logDate)

		var prevBalanceAfter *float64
		err = tx.QueryRow(ctx, `SELECT BALANCE_AFTER FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 AND CREATED_AT = $2`, userID, createdAt).Scan(&prevBalanceAfter)
		if err != nil && err.Error() != "no rows in result set" {
			return 0, err
		}
		var prevBalance float64 = 0
		if prevBalanceAfter != nil {
			prevBalance = *prevBalanceAfter
		}

		txBalanceAfter := prevBalance + oldTotalCost
		// ensure refund is sent after delivery record
		createdAt = createdAt.Add(1 * time.Second)
		_, err = tx.Exec(ctx, `
			INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, CREATED_AT)
			VALUES ($1, 'refund', 'confirmed', $2, $3, $4)
		`, userID, oldTotalCost, txBalanceAfter, createdAt)

		if err != nil {
			return 0, err
		}
		err = utils.RecalculateBalances(ctx, tx, utils.REFUND, userID, createdAt, oldTotalCost)
		if err != nil {
			return 0, err
		}

		txBalanceAfter -= newTotalCost
		// ensure new delivery is sent after refund record
		createdAt = createdAt.Add(1 * time.Second)
		_, err = tx.Exec(ctx, `
			INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, CREATED_AT)
			VALUES ($1, 'delivery', 'confirmed', $2, $3, $4)
		`, userID, newTotalCost, txBalanceAfter, createdAt)

		if err != nil {
			return 0, err
		}
		err = utils.RecalculateBalances(ctx, tx, utils.DELIVERY, userID, createdAt, oldTotalCost)
		if err != nil {
			return 0, err
		}

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
