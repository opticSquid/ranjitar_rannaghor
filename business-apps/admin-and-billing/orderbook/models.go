package orderbook

import "time"

type MealType string

const (
	LUNCH  MealType = "lunch"
	DINNER MealType = "dinner"
)

type OrderDetails struct {
	ItemId   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type EntryRequest struct {
	UserID       int            `json:"user_id"`
	EntryDate    time.Time      `json:"log_date"`
	MealType     MealType       `json:"meal_type"`
	OrderDetails []OrderDetails `json:"order_details"`
}

type MenuItemCategory string

const (
	COMBO_THALI MenuItemCategory = "combo_thali"
	A_LA_CARTE  MenuItemCategory = "a_la_carte"
)

type MenuItemResponse struct {
	ItemId        int              `json:"item_id"`
	ItemName      string           `json:"item_name"`
	Category      MenuItemCategory `json:"category"`
	IsActive      bool             `json:"is_active"`
	LatestPrice   float64          `json:"latest_price"`
	EffectiveFrom time.Time        `json:"effective_from"`
}
