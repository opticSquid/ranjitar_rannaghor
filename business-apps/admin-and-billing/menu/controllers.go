package menu

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var m NewMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	menuItem, err := createMenuItemService(r.Context(), m)
	if err != nil {
		if errors.Is(err, ErrInvalidMenuCategory) || errors.Is(err, ErrEffectiveFromValueOfPast) || errors.Is(err, ErrMenuItemExists) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(menuItem)
}

func GetMenuItems(w http.ResponseWriter, r *http.Request) {
	menuItems, err := getMenuItemsService(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(menuItems)
}

func UpdateMenuItemDetails(w http.ResponseWriter, r *http.Request) {

	var req MenuItemUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedMenuItem, err := updateMenuItemService(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedMenuItem)
}

func UpdateMenuItemPrice(w http.ResponseWriter, r *http.Request) {
	var req PriceUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedMenuItem, err := updateMenuItemPriceService(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEffectiveFromValueOfPast) || errors.Is(err, ErrMenuItemDoesNotExist) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedMenuItem)
}

func DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	itemID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := deleteMenuItemService(r.Context(), itemID); err != nil {
		if errors.Is(err, ErrMenuItemDoesNotExist) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	itemId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	entries, err := getPriceHistoryService(r.Context(), itemId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(entries)
}
