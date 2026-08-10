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
)

type NewMenuItemRequest struct {
	Name          string    `json:"item_name"`
	Category      string    `json:"category"`
	Price         float64   `json:"price"`
	EffectiveFrom time.Time `json:"effective_from"`
}

type MenuItemResponse struct {
	ItemId        int              `json:"item_id"`
	Name          string           `json:"item_name"`
	Category      MenuItemCategory `json:"category"`
	LatestPrice   float64          `json:"latest_price"`
	EffectiveFrom time.Time        `json:"effective_from"`
}

type MenuItem struct {
	itemId    int
	name      string
	category  MenuItemCategory
	isActive  bool
	createdAt time.Time
}

// MenuPriceHistory represents a historical price for a menu item
type MenuPriceHistory struct {
	priceId       int
	itemId        int
	price         float64
	effectiveFrom time.Time
	createdAt     time.Time
}
