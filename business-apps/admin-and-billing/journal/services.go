package journal

import (
	"context"
	"fmt"
	"time"
	_ "time/tzdata"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/meals"
)

func getCreationTime(logDate time.Time) time.Time {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Error loading location:", err)
	}
	logDate = logDate.In(loc)
	y, m, d := logDate.Date()
	now := time.Now().In(loc)
	created_at := time.Date(y, m, d, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), loc)
	return created_at.UTC()
}

func constructCreationTime(dateVar time.Time, timeVar time.Time) time.Time {
	y, m, d := dateVar.Date()
	h, M, s := timeVar.Clock()
	return time.Date(y, m, d, h, M, s, 0, timeVar.Location())
}

func CalculateTotalCost(log EntryRequest, prices map[string]float64) float64 {
	mealPrice := 0.0
	if log.HasMainMeal {
		mealPrice = prices["standard"]
		if log.IsSpecial {
			mealPrice = prices["special"]
		}
	}
	totalCost := mealPrice + (float64(log.ExtraRiceQty) * prices["rice"]) + (float64(log.ExtraRotiQty) * prices["roti"]) + (float64(log.ExtraChickenQty) * prices["chicken"]) + (float64(log.ExtraFishQty) * prices["fish"]) + (float64(log.ExtraEggQty) * prices["egg"]) + (float64(log.ExtraVegetableQty) * prices["vegetable"])
	return totalCost
}

func CreateDailyEntryService(ctx context.Context, entry EntryRequest) error {
	var menu_items []string
	if entry.IsSpecialMenu {
		menu_items = append(menu_items, "special_menu")
	} else {
		menu_items = append(menu_items, entry.MainCourseName)
	}
	if entry.ExtraRiceQty > 0 {
		menu_items = append(menu_items, "extra_rice")
	}
	if entry.ExtraRotiQty > 0 {
		menu_items = append(menu_items, "extra_roti")
	}
	if entry.ExtraVegetableQty > 0 {
		menu_items = append(menu_items, "extra_vegetable")
	}
	if entry.ExtraEggQty > 0 {
		menu_items = append(menu_items, "extra_egg")
	}
	if entry.ExtraFishQty > 0 {
		menu_items = append(menu_items, "extra_fish")
	}
	if entry.ExtraChickenQty > 0 {
		menu_items = append(menu_items, "extra_chicken")
	}

	// todo: this should be like: item_name: {item_id, price}
	prices, err := meals.FetchMealPricesInternal(ctx, entry.EntryDate, menu_items)
	if err != nil {
		return err
	}
	wallet_txns := make([]walletTxn, len(prices))
	for _, itm := range menu_items {
		qty := 1
		if itm == "extra_rice" {
			qty = entry.ExtraRiceQty
		}
		if entry.ExtraRotiQty > 0 {
			qty = entry.ExtraRotiQty
		}
		if entry.ExtraVegetableQty > 0 {
			qty = entry.ExtraVegetableQty
		}
		if entry.ExtraEggQty > 0 {
			qty = entry.ExtraEggQty
		}
		if entry.ExtraFishQty > 0 {
			qty = entry.ExtraFishQty
		}
		if entry.ExtraChickenQty > 0 {
			qty = entry.ExtraChickenQty
		}
		wallet_txns = append(wallet_txns, walletTxn{nil, entry.UserID, entry.EntryDate, "delivery", entry.MealType, itm, qty})
	}

	return InsertWalletTxn(ctx, entry, totalCost, createdAt)
}

func DeleteDailyEntryService(ctx context.Context, logID int) (float64, error) {
	return DeleteDailyEntryFromDB(ctx, logID)
}

func UpdateDailyEntryService(ctx context.Context, logID int, req EntryRequest) (float64, error) {
	// Let repository compute new total cost using the original creation timestamp
	return UpdateDailyEntryInDB(ctx, logID, req)
}

func GetDailyEntriesService(ctx context.Context, date time.Time, userID int) ([]DailyLog, error) {
	return FetchDailyEntries(ctx, date, userID)
}
