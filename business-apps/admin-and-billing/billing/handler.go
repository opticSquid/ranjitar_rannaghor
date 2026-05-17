package billing

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/model"
)

func GetBill(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	userID, _ := strconv.Atoi(userIDStr)
	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)

	var report model.BillReport
	report.StartDate = startDate
	report.EndDate = endDate

	// Get User Info
	dbPool := database.GetDbConn()
	err := dbPool.QueryRow(r.Context(), `
		SELECT u.USER_ID, u.NAME, u.MOBILE_NO, u.BUILDING_NO, u.ROOM_NO, u.ROLE, u.PLAN
		FROM USERS u
		WHERE u.USER_ID = $1
	`, userID).Scan(&report.User.UserID, &report.User.Name, &report.User.MobileNo, &report.User.BuildingNo, &report.User.RoomNo, &report.User.Role, &report.User.Plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get Logs
	rows, err := dbPool.Query(r.Context(), `
		SELECT LOG_ID, LOG_DATE, MEAL_TYPE, HAS_MAIN_MEAL, IS_SPECIAL, SPECIAL_DISH_NAME, EXTRA_RICE_QTY, EXTRA_ROTI_QTY, TOTAL_COST
		FROM DAILY_LOGS
		WHERE USER_ID = $1 AND LOG_DATE BETWEEN $2 AND $3
		ORDER BY LOG_DATE ASC, MEAL_TYPE DESC
	`, userID, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var l model.DailyLog
		err := rows.Scan(&l.LogID, &l.LogDate, &l.MealType, &l.HasMainMeal, &l.IsSpecial, &l.SpecialDishName, &l.ExtraRiceQty, &l.ExtraRotiQty, &l.TotalCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		report.Logs = append(report.Logs, l)
		report.TotalSpent += l.TotalCost
	}

	err = dbPool.QueryRow(r.Context(), `SELECT BALANCE_AFTER
	FROM WALLET_TRANSACTIONS
	WHERE USER_ID = $1
	AND STATUS = 'confirmed'
	AND CREATED_AT <= $2
	ORDER BY CREATED_AT DESC
	LIMIT 1`, userID, endDate).Scan(&report.ClosingBalance)

	// Closing balance is current balance
	report.User.Balance = report.ClosingBalance

	// Calculate total recharges during billing period
	dbPool.QueryRow(r.Context(), `
		SELECT COALESCE(SUM(AMOUNT), 0)
		FROM WALLET_TRANSACTIONS
		WHERE USER_ID = $1
		  AND TXN_TYPE = 'recharge'
		  AND STATUS = 'confirmed'
		  AND CREATED_AT >= $2
		  AND CREATED_AT <= $3
	`, userID, startDate, endDate.AddDate(0, 0, 1)).Scan(&report.TotalRecharges)

	// Opening balance = Balance before the billing period started
	// Get the BALANCE_AFTER from the last confirmed transaction before the start date
	// If no transactions exist before start date, opening balance is 0
	var openingBalance *float64
	err = dbPool.QueryRow(r.Context(), `
		SELECT BALANCE_AFTER
		FROM WALLET_TRANSACTIONS
		WHERE USER_ID = $1
		  AND STATUS = 'confirmed'
		  AND CREATED_AT < $2
		ORDER BY CREATED_AT DESC, TXN_ID DESC
		LIMIT 1
	`, userID, startDate).Scan(&openingBalance)

	if err != nil || openingBalance == nil {
		// No transactions before start date, opening balance is 0
		report.OpeningBalance = 0
	} else {
		report.OpeningBalance = *openingBalance
	}

	json.NewEncoder(w).Encode(report)
}
