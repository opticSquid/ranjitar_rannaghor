package billing

import (
	"context"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/journal"
)

func FetchBillReportFromDB(ctx context.Context, userID int, startDate, endDate time.Time) (BillReport, error) {
	var report BillReport
	report.StartDate = startDate
	report.EndDate = endDate

	dbPool := database.GetDbConn()
	err := dbPool.QueryRow(ctx, `
		SELECT u.USER_ID, u.NAME, u.MOBILE_NO, u.BUILDING_NO, u.ROOM_NO, u.ROLE, u.PLAN
		FROM USERS u
		WHERE u.USER_ID = $1
	`, userID).Scan(&report.User.UserID, &report.User.Name, &report.User.MobileNo, &report.User.BuildingNo, &report.User.RoomNo, &report.User.Role, &report.User.Plan)
	if err != nil {
		return report, err
	}

	rows, err := dbPool.Query(ctx, `
		SELECT LOG_ID, LOG_DATE, MEAL_TYPE, HAS_MAIN_MEAL, IS_SPECIAL, SPECIAL_DISH_NAME, EXTRA_RICE_QTY, EXTRA_ROTI_QTY, TOTAL_COST
		FROM DAILY_LOGS
		WHERE USER_ID = $1 AND LOG_DATE BETWEEN $2 AND $3
		ORDER BY LOG_DATE ASC, MEAL_TYPE DESC
	`, userID, startDate, endDate)
	if err != nil {
		return report, err
	}
	defer rows.Close()

	for rows.Next() {
		var l journal.DailyLog
		err := rows.Scan(&l.LogID, &l.LogDate, &l.MealType, &l.HasMainMeal, &l.IsSpecial, &l.SpecialDishName, &l.ExtraRiceQty, &l.ExtraRotiQty, &l.TotalCost)
		if err != nil {
			return report, err
		}
		report.Logs = append(report.Logs, l)
		report.TotalSpent += l.TotalCost
	}

	err = dbPool.QueryRow(ctx, `SELECT BALANCE_AFTER
	FROM WALLET_TRANSACTIONS
	WHERE USER_ID = $1
	AND STATUS = 'confirmed'
	AND CREATED_AT <= $2
	ORDER BY CREATED_AT DESC
	LIMIT 1`, userID, endDate).Scan(&report.ClosingBalance)
	// original code did not check error here

	report.User.Balance = report.ClosingBalance

	dbPool.QueryRow(ctx, `
		SELECT COALESCE(SUM(AMOUNT), 0)
		FROM WALLET_TRANSACTIONS
		WHERE USER_ID = $1
		  AND TXN_TYPE = 'recharge'
		  AND STATUS = 'confirmed'
		  AND CREATED_AT >= $2
		  AND CREATED_AT <= $3
	`, userID, startDate, endDate.AddDate(0, 0, 1)).Scan(&report.TotalRecharges)

	var openingBalance *float64
	err = dbPool.QueryRow(ctx, `
		SELECT BALANCE_AFTER
		FROM WALLET_TRANSACTIONS
		WHERE USER_ID = $1
		  AND STATUS = 'confirmed'
		  AND CREATED_AT < $2
		ORDER BY CREATED_AT DESC, TXN_ID DESC
		LIMIT 1
	`, userID, startDate).Scan(&openingBalance)

	if err != nil || openingBalance == nil {
		report.OpeningBalance = 0
	} else {
		report.OpeningBalance = *openingBalance
	}

	return report, nil
}
