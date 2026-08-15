package journal

import "time"

type EntryRequest struct {
	UserID            int       `json:"user_id"`
	EntryDate         time.Time `json:"log_date"`
	MealType          string    `json:"meal_type"`
	MainCourseName    string    `json:"main_course_name"`
	IsSpecialMenu     bool      `json:"is_special_menu"`
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

type walletTxn struct {
	txnId       int
	userId      int
	txnDate     time.Time
	txnType     string
	mealType    string
	dishName    string
	quantity    string
	amount      float64
	referenceId string
	menuItemId  int
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
