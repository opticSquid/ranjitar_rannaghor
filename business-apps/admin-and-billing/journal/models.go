package journal

import "time"

type EntryRequest struct {
	UserID            int       `json:"user_id"`
	LogDate           time.Time `json:"log_date"`
	MealType          string    `json:"meal_type"`
	HasMainMeal       bool      `json:"has_main_meal"`
	IsSpecial         bool      `json:"is_special"`
	SpecialDishName   string    `json:"special_dish_name"`
	ExtraRiceQty      int       `json:"extra_rice_qty"`
	ExtraRotiQty      int       `json:"extra_roti_qty"`
	ExtraChickenQty   int       `json:"extra_chicken_qty"`
	ExtraFishQty      int       `json:"extra_fish_qty"`
	ExtraEggQty       int       `json:"extra_egg_qty"`
	ExtraVegetableQty int       `json:"extra_vegetable_qty"`
}

type DailyLog struct {
	LogID             int       `json:"log_id"`
	UserID            int       `json:"user_id"`
	UserName          string    `json:"user_name,omitempty"`
	LogDate           time.Time `json:"log_date"`
	MealType          string    `json:"meal_type"`
	HasMainMeal       bool      `json:"has_main_meal"`
	IsSpecial         bool      `json:"is_special"`
	SpecialDishName   string    `json:"special_dish_name"`
	ExtraRiceQty      int       `json:"extra_rice_qty"`
	ExtraRotiQty      int       `json:"extra_roti_qty"`
	ExtraChickenQty   int       `json:"extra_chicken_qty"`
	ExtraFishQty      int       `json:"extra_fish_qty"`
	ExtraEggQty       int       `json:"extra_egg_qty"`
	ExtraVegetableQty int       `json:"extra_vegetable_qty"`
	TotalCost         float64   `json:"total_cost"`
}
