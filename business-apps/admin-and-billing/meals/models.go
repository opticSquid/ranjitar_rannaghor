package meals

import "time"

type MealPrice struct {
	ItemID    string    `json:"item_id"`
	ItemName  string    `json:"item_name"`
	Price     float64   `json:"price"`
	UpdatedAt time.Time `json:"updated_at"`
}
