package journal

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/soumalya/food-delivery-admin/database"
	"github.com/soumalya/food-delivery-admin/meals"
	"github.com/soumalya/food-delivery-admin/model"
	"github.com/soumalya/food-delivery-admin/utils"
)

// CalculateTotalCost computes the total cost for an EntryRequest using the provided prices map.
// It mirrors the calculation logic used by the handlers and is exported so it can be unit-tested.
func CalculateTotalCost(log model.EntryRequest, prices map[string]float64) float64 {
	mealPrice := 0.0
	if log.HasMainMeal {
		mealPrice = prices["standard"]
		if log.IsSpecial {
			mealPrice = prices["special"]
		}
	}
	totalCost := mealPrice + (float64(log.ExtraRiceQty) * prices["rice"]) + (float64(log.ExtraRotiQty) * prices["roti"]) + (float64(log.ExtraChickenQty) * prices["chicken"]) + (float64(log.ExtraFishQty) * prices["fish"]) + (float64(log.ExtraEggQty) * prices["egg"]) + (float64(log.ExtraVegetableQty) * prices["vegetable"])
	return totalCost
}

func getCreationTime(logDate time.Time) string {
	yyyy, MM, dd := logDate.Date()
	utc_time := time.Now().UTC()
	return strconv.Itoa(yyyy) + "-" + strconv.Itoa(int(MM)) + "-" + strconv.Itoa(dd) + " " + strconv.Itoa(utc_time.Hour()) + ":" + strconv.Itoa(utc_time.Minute()) + ":" + strconv.Itoa(utc_time.Second()) + "." + strconv.Itoa(utc_time.Nanosecond()) + "+" + "00"
}

func CreateDailyEntry(w http.ResponseWriter, r *http.Request) {
	var log model.EntryRequest
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Get current meal prices
	prices := meals.GetMealPricesInternal(r.Context())

	// Calculate cost
	totalCost := CalculateTotalCost(log, prices)

	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Insert Log
	_, err = tx.Exec(r.Context(), `
		INSERT INTO DAILY_LOGS (USER_ID, LOG_DATE, MEAL_TYPE, HAS_MAIN_MEAL, IS_SPECIAL, SPECIAL_DISH_NAME, EXTRA_RICE_QTY, EXTRA_ROTI_QTY, EXTRA_CHICKEN_QTY, EXTRA_FISH_QTY, EXTRA_EGG_QTY, EXTRA_VEGETABLE_QTY, TOTAL_COST)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9	, $10, $11, $12, $13)
	`, log.UserID, log.LogDate, log.MealType, log.HasMainMeal, log.IsSpecial, log.SpecialDishName, log.ExtraRiceQty, log.ExtraRotiQty, log.ExtraChickenQty, log.ExtraFishQty, log.ExtraEggQty, log.ExtraVegetableQty, totalCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch previous wallet balance
	// Use the construct created at using log date and current time
	createdAt := getCreationTime(log.LogDate)

	var prevBalanceAfter *float64
	err = tx.QueryRow(r.Context(), `SELECT BALANCE_AFTER FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 AND CREATED_AT < $2 ORDER BY CREATED_AT DESC LIMIT 1`, log.UserID, createdAt).Scan(&prevBalanceAfter)

	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var currentBalance float64 = 0
	if prevBalanceAfter != nil {
		currentBalance = *prevBalanceAfter
	}
	newBalance := currentBalance - totalCost

	_, err = tx.Exec(r.Context(), `
		INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, CREATED_AT)
		VALUES ($1, 'delivery', 'confirmed', $2, $3, $4)
	`, log.UserID, totalCost, newBalance, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Recalculate all future balances for this user
	err = utils.RecalculateBalances(r.Context(), tx, log.UserID, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx.Commit(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"new_balance": newBalance})
}

func DeleteDailyEntry(w http.ResponseWriter, r *http.Request) {
	logIDStr := chi.URLParam(r, "id")
	logID, _ := strconv.Atoi(logIDStr)

	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Get info to refund
	var userID int
	var totalCost float64
	var logDate time.Time
	err = tx.QueryRow(r.Context(), `SELECT USER_ID, TOTAL_COST, LOG_DATE FROM DAILY_LOGS WHERE LOG_ID = $1`, logID).Scan(&userID, &totalCost, &logDate)
	if err != nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	// Delete
	_, err = tx.Exec(r.Context(), `DELETE FROM DAILY_LOGS WHERE LOG_ID = $1`, logID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log Wallet Transaction — use the original log timestamp directly
	createdAt := getCreationTime(logDate)

	// Fetch previous wallet balance before createdAt
	var prevBalanceAfter *float64
	err = tx.QueryRow(r.Context(), `SELECT BALANCE_AFTER FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 AND CREATED_AT < $2 ORDER BY CREATED_AT DESC LIMIT 1`, userID, createdAt).Scan(&prevBalanceAfter)

	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var currentBalance float64 = 0
	if prevBalanceAfter != nil {
		currentBalance = *prevBalanceAfter
	}
	// Refund should increase the balance
	newBalance := currentBalance + totalCost

	_, err = tx.Exec(r.Context(), `
		INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, CREATED_AT)
		VALUES ($1, 'refund', 'confirmed', $2, $3, $4)
	`, userID, totalCost, newBalance, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Recalculate all future balances for this user
	err = utils.RecalculateBalances(r.Context(), tx, userID, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx.Commit(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"new_balance": newBalance})
}

func UpdateDailyEntry(w http.ResponseWriter, r *http.Request) {
	logIDStr := chi.URLParam(r, "id")
	logID, _ := strconv.Atoi(logIDStr)

	var req model.EntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Get current meal prices
	prices := meals.GetMealPricesInternal(r.Context())

	// Calculate new cost
	newTotalCost := CalculateTotalCost(req, prices)

	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Get old info (including LOG_DATE for timestamping the txn)
	var userID int
	var oldTotalCost float64
	var logDate time.Time
	err = tx.QueryRow(r.Context(), `SELECT USER_ID, TOTAL_COST, LOG_DATE FROM DAILY_LOGS WHERE LOG_ID = $1`, logID).Scan(&userID, &oldTotalCost, &logDate)
	if err != nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	// Update Log
	_, err = tx.Exec(r.Context(), `
		UPDATE DAILY_LOGS
		SET MEAL_TYPE = $1, HAS_MAIN_MEAL = $2, IS_SPECIAL = $3, SPECIAL_DISH_NAME = $4, EXTRA_RICE_QTY = $5, EXTRA_ROTI_QTY = $6, TOTAL_COST = $7
		WHERE LOG_ID = $8
	`, req.MealType, req.HasMainMeal, req.IsSpecial, req.SpecialDishName, req.ExtraRiceQty, req.ExtraRotiQty, newTotalCost, logID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Adjust Wallet using WALLET_TRANSACTIONS only (WALLET table is deprecated)
	costDiff := newTotalCost - oldTotalCost
	if costDiff != 0 {
		// Determine transaction type and amount
		txnType := "delivery"
		txnAmount := costDiff
		if costDiff < 0 {
			txnType = "refund"
			txnAmount = -costDiff
		}

		// Use the original log timestamp for the txn
		createdAt := getCreationTime(logDate)

		// Fetch previous per-transaction balance as of createdAt
		var prevBalanceAfter *float64
		err = tx.QueryRow(r.Context(), `SELECT BALANCE_AFTER FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 AND CREATED_AT < $2 ORDER BY CREATED_AT DESC, TXN_ID DESC LIMIT 1`, userID, createdAt).Scan(&prevBalanceAfter)
		if err != nil && err.Error() != "no rows in result set" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var prevBalance float64 = 0
		if prevBalanceAfter != nil {
			prevBalance = *prevBalanceAfter
		}

		// Balance after applying this diff at createdAt. For costDiff > 0 we deduct, for costDiff < 0 we add.
		txBalanceAfter := prevBalance - costDiff

		_, err = tx.Exec(r.Context(), `
			INSERT INTO WALLET_TRANSACTIONS (USER_ID, TXN_TYPE, STATUS, AMOUNT, BALANCE_AFTER, CREATED_AT)
			VALUES ($1, $2, 'confirmed', $3, $4, $5)
		`, userID, txnType, txnAmount, txBalanceAfter, createdAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Recalculate all future balances for this user
		err = utils.RecalculateBalances(r.Context(), tx, userID, createdAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Return the latest per-transaction balance after the operation
	var finalBalance float64
	err = tx.QueryRow(r.Context(), `SELECT COALESCE(BALANCE_AFTER, 0) FROM WALLET_TRANSACTIONS WHERE USER_ID = $1 ORDER BY CREATED_AT DESC, TXN_ID DESC LIMIT 1`, userID).Scan(&finalBalance)
	if err != nil {
		// If there are no transactions, default to 0
		if err.Error() == "no rows in result set" {
			finalBalance = 0
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tx.Commit(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"new_balance": finalBalance})
}

func GetDailyEntries(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	userIDStr := r.URL.Query().Get("user_id")

	if dateStr == "" {
		http.Error(w, "Date is required", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	query := `
		SELECT l.LOG_ID, l.USER_ID, u.NAME as USER_NAME, l.LOG_DATE, l.MEAL_TYPE,
		       l.HAS_MAIN_MEAL, l.IS_SPECIAL, l.SPECIAL_DISH_NAME,
		       l.EXTRA_RICE_QTY, l.EXTRA_ROTI_QTY, l.EXTRA_CHICKEN_QTY, l.EXTRA_FISH_QTY, l.EXTRA_EGG_QTY, l.EXTRA_VEGETABLE_QTY, l.TOTAL_COST
		FROM DAILY_LOGS l
		JOIN USERS u ON l.USER_ID = u.USER_ID
		WHERE l.LOG_DATE = $1
	`
	args := []interface{}{date}

	if userIDStr != "" && userIDStr != "0" {
		userID, _ := strconv.Atoi(userIDStr)
		query += " AND l.USER_ID = $2"
		args = append(args, userID)
	}

	query += " ORDER BY u.NAME ASC, l.MEAL_TYPE DESC"

	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []model.DailyLog
	for rows.Next() {
		var l model.DailyLog
		err := rows.Scan(&l.LogID, &l.UserID, &l.UserName, &l.LogDate, &l.MealType, &l.HasMainMeal, &l.IsSpecial, &l.SpecialDishName, &l.ExtraRiceQty, &l.ExtraRotiQty, &l.ExtraChickenQty, &l.ExtraFishQty, &l.ExtraEggQty, &l.ExtraVegetableQty, &l.TotalCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logs = append(logs, l)
	}

	json.NewEncoder(w).Encode(logs)
}
