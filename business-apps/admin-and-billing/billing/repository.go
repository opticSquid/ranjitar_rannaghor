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

	// Fetch daily logs (LOG_DATE is assumed to be date-only, so BETWEEN startDate and endDate is inclusive)
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

	// For timestamp-based wallet transactions we treat endDate as inclusive by using < endDate + 1 day
	endDateExclusive := endDate.AddDate(0, 0, 1)

	// Closing balance: latest confirmed wallet transaction at or before end date (inclusive of the day)
	if err := dbPool.QueryRow(ctx, `SELECT BALANCE_AFTER
		FROM WALLET_TRANSACTIONS
		WHERE USER_ID = $1
		AND STATUS = 'confirmed'
		AND CREATED_AT < $2
		ORDER BY CREATED_AT DESC
		LIMIT 1`, userID, endDateExclusive).Scan(&report.ClosingBalance); err != nil {
		// if no rows found or other errors, default to 0
		report.ClosingBalance = 0
	}

	report.User.Balance = report.ClosingBalance

	// Total recharges in the inclusive date range
	if err := dbPool.QueryRow(ctx, `
		SELECT COALESCE(SUM(AMOUNT), 0)
		FROM WALLET_TRANSACTIONS
		WHERE USER_ID = $1
		  AND TXN_TYPE = 'recharge'
		  AND STATUS = 'confirmed'
		  AND CREATED_AT >= $2
		  AND CREATED_AT < $3
	`, userID, startDate, endDateExclusive).Scan(&report.TotalRecharges); err != nil {
		report.TotalRecharges = 0
	}

	// Opening balance: most recent confirmed wallet transaction strictly before start date
	if err := dbPool.QueryRow(ctx, `
		SELECT BALANCE_AFTER
		FROM WALLET_TRANSACTIONS
		WHERE USER_ID = $1
		  AND STATUS = 'confirmed'
		  AND CREATED_AT < $2
		ORDER BY CREATED_AT DESC, TXN_ID DESC
		LIMIT 1
	`, userID, startDate).Scan(&report.OpeningBalance); err != nil {
		// default to 0 if not found
		report.OpeningBalance = 0
	}

	return report, nil
}
