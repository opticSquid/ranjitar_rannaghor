package meals

import (
	"encoding/json"
	"net/http"
	"time"

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

// PriceUpdateRequest is the payload for creating a price history entry
type PriceUpdateRequest struct {
	Price         float64 `json:"price"`
	EffectiveFrom string  `json:"effective_from"` // accept local datetime or RFC3339 string
	CreatedBy     string  `json:"created_by"`
}

// CreatePrice registers a new price with an effective timestamp for an item
func CreatePrice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req PriceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var eff time.Time
	var err error
	if req.EffectiveFrom == "" {
		eff = time.Now().UTC()
	} else {
		// Try parsing RFC3339 first
		eff, err = time.Parse(time.RFC3339, req.EffectiveFrom)
		if err != nil {
			// Try parsing datetime-local format without timezone: 2006-01-02T15:04
			effLocal, err2 := time.ParseInLocation("2006-01-02T15:04", req.EffectiveFrom, time.Local)
			if err2 != nil {
				http.Error(w, "invalid effective_from format", http.StatusBadRequest)
				return
			}
			eff = effLocal.UTC()
		} else {
			// parsed successfully as RFC3339; normalize to UTC
			eff = eff.UTC()
		}
	}

	if err := InsertPriceHistory(r.Context(), id, req.Price, eff, req.CreatedBy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If effective_from is now or in the past, update current MEAL_PRICES for convenience
	if !eff.After(time.Now().UTC()) {
		// best-effort update
		_ = UpdateMealPriceInDB(r.Context(), id, req.Price)
	}

	w.WriteHeader(http.StatusCreated)
}

// GetPriceHistory returns the price history for an item
func GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entries, err := FetchPriceHistoryForItem(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(entries)
}
