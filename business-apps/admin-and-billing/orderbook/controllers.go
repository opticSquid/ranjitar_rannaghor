package orderbook

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req NewOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	order, err := createOrderService(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUserDoesNotExist) || errors.Is(err, ErrInvalidMealType) || errors.Is(err, ErrMenuItemDoesNotExist) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orderId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := deleteOrderService(r.Context(), orderId); err != nil {
		if errors.Is(err, ErrOrderDoesNotExist) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func UpdateDailyEntry(w http.ResponseWriter, r *http.Request) {
	logIDStr := chi.URLParam(r, "id")
	logID, _ := strconv.Atoi(logIDStr)

	var req NewOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newBalance, err := updateOrderService(r.Context(), logID, req)
	if err != nil {
		if err.Error() == "Entry not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"new_balance": newBalance})
}
