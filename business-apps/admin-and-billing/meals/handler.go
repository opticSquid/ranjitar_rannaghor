package meals

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/model"
)

func GetMealPricesInternal(ctx context.Context) map[string]float64 {
	prices := make(map[string]float64)
	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(ctx, "SELECT ITEM_ID, PRICE FROM MEAL_PRICES")
	if err != nil {
		slog.Error("Failed to get meal prices", "err", err)
		return prices
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var price float64
		if err := rows.Scan(&id, &price); err == nil {
			prices[id] = price
		}
	}
	return prices
}

func CreateMeal(w http.ResponseWriter, r *http.Request) {
	var m model.MealPrice
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dbPool := database.GetDbConn()
	err := dbPool.QueryRow(r.Context(), `
		INSERT INTO MEAL_PRICES (ITEM_NAME, ITEM_PRICE)
		VALUES ($1, $2)
		RETURNING ITEM_ID
	`, m.ItemName, m.Price).Scan(&m.ItemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(m)
}

func GetMeals(w http.ResponseWriter, r *http.Request) {
	dbPool := database.GetDbConn()
	rows, err := dbPool.Query(r.Context(), "SELECT ITEM_ID, ITEM_NAME, PRICE, UPDATED_AT FROM MEAL_PRICES ORDER BY PRICE DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var prices []model.MealPrice
	for rows.Next() {
		var p model.MealPrice
		err := rows.Scan(&p.ItemID, &p.ItemName, &p.Price, &p.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		prices = append(prices, p)
	}

	json.NewEncoder(w).Encode(prices)
}

func UpdateMeal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var p model.MealPrice
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbPool := database.GetDbConn()
	_, err := dbPool.Exec(r.Context(), `
		UPDATE MEAL_PRICES SET PRICE = $1, UPDATED_AT = CURRENT_TIMESTAMP
		WHERE ITEM_ID = $2
	`, p.Price, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteMeal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var p model.MealPrice
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dbPool := database.GetDbConn()
	_, err := dbPool.Exec(r.Context(), `
		DELETE FROM MEAL_PRICES
		WHERE ITEM_ID = $1
	`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
