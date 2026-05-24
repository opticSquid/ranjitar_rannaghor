package meals

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func CreateMeal(w http.ResponseWriter, r *http.Request) {
	var m MealPrice
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := CreateMealService(r.Context(), &m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(m)
}

func GetMeals(w http.ResponseWriter, r *http.Request) {
	prices, err := GetMealsService(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(prices)
}

func UpdateMeal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var p MealPrice
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := UpdateMealService(r.Context(), id, p.Price); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteMeal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := DeleteMealService(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
