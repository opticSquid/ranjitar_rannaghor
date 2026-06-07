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

func CreateDailyEntryService(ctx context.Context, log EntryRequest) (float64, error) {
	createdAt := getCreationTime(log.LogDate)
	prices := meals.GetMealPricesAt(ctx, createdAt)

	totalCost := CalculateTotalCost(log, prices)

	return CreateDailyEntryInDB(ctx, log, totalCost, createdAt)
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
