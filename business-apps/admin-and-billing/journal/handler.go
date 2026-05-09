package journal

import (
	"encoding/json"
	"net/http"

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
	// Use the provided log timestamp directly as createdAt
	createdAt := log.LogDate

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
