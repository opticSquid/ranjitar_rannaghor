package meals

import "time"

type MealPrice struct {
	ItemID    string    `json:"item_id"`
	ItemName  string    `json:"item_name"`
	Price     float64   `json:"price"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PriceHistoryEntry represents a historical price for a menu item
type PriceHistoryEntry struct {
	ID            int       `json:"id"`
	ItemID        string    `json:"item_id"`
	Price         float64   `json:"price"`
	EffectiveFrom time.Time `json:"effective_from"`
	CreatedBy     *string   `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
