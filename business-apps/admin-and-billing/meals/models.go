package meals

import (
	"errors"
	"time"
)

type MenuItemCategory string

const (
	COMBO_THALI MenuItemCategory = "combo_thali"
	A_LA_CARTE  MenuItemCategory = "a_la_carte"
)

var (
	ErrInvalidMenuCategory      = errors.New("menu item category in invalid")
	ErrEffectiveFromValueOfPast = errors.New("effective_from value is in the past of current time")
	ErrMenuItemExists           = errors.New("menu item with the same name already exists")
	ErrMenuItemDoesNotExist     = errors.New("menu item does not exist")
)

type NewMenuItemRequest struct {
	Name          string    `json:"item_name"`
	Category      string    `json:"category"`
	Price         float64   `json:"price"`
	EffectiveFrom time.Time `json:"effective_from"`
}

type MenuItemResponse struct {
	ItemId        int              `json:"item_id"`
	ItemName      string           `json:"item_name"`
	Category      MenuItemCategory `json:"category"`
	IsActive      bool             `json:"is_active"`
	LatestPrice   float64          `json:"latest_price"`
	EffectiveFrom time.Time        `json:"effective_from"`
}

type MenuItemUpdateRequest struct {
	ItemId   int    `json:"item_id"`
	Name     string `json:"item_name"`
	Category string `json:"category"`
	IsActive bool   `json:"is_active"`
}

type menuItem struct {
	itemId    int
	itemName  string
	category  MenuItemCategory
	isActive  bool
	createdAt time.Time
}

type PriceUpdateRequest struct {
	ItemId        int       `json:"item_id"`
	NewPrice      float64   `json:"new_price"`
	EffectiveFrom time.Time `json:"effective_from"` // accept local datetime or RFC3339 string
}

type PriceUpdateResponse struct {
	ItemId        int       `json:"item_id"`
	Price         float64   `json:"price"`
	EffectiveFrom time.Time `json:"effective_from"`
}

type PriceHistoryResponse struct {
	Price         float64   `json:"price"`
	EffectiveFrom time.Time `json:"effective_from"`
}

type menuPriceSchedule struct {
	priceId       int
	itemId        int
	price         float64
	effectiveFrom time.Time
	createdAt     time.Time
}
