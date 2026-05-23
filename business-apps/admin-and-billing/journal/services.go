package journal

import (
	"context"
	"strconv"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/meals"
)

func getCreationTime(logDate time.Time) time.Time {
	yyyy, MM, dd := logDate.Date()
	utc_time := time.Now().UTC()
	crtAtTime := strconv.Itoa(yyyy) + "-" + strconv.Itoa(int(MM)) + "-" + strconv.Itoa(dd) + " " + strconv.Itoa(utc_time.Hour()) + ":" + strconv.Itoa(utc_time.Minute()) + ":" + strconv.Itoa(utc_time.Second()) + "." + strconv.Itoa(utc_time.Nanosecond()) + "+" + "00"
	layout := "2006-01-02 15:04:05.000000-07"
	t, err := time.Parse(layout, crtAtTime)
	if err != nil {
		return time.Now()
	}
	return t
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
	prices := meals.GetMealPricesInternal(ctx)
	totalCost := CalculateTotalCost(log, prices)
	createdAt := getCreationTime(log.LogDate)

	return CreateDailyEntryInDB(ctx, log, totalCost, createdAt)
}

func DeleteDailyEntryService(ctx context.Context, logID int) (float64, error) {
	return DeleteDailyEntryFromDB(ctx, logID)
}

func UpdateDailyEntryService(ctx context.Context, logID int, req EntryRequest) (float64, error) {
	prices := meals.GetMealPricesInternal(ctx)
	newTotalCost := CalculateTotalCost(req, prices)

	return UpdateDailyEntryInDB(ctx, logID, req, newTotalCost)
}

func GetDailyEntriesService(ctx context.Context, date time.Time, userID int) ([]DailyLog, error) {
	return FetchDailyEntries(ctx, date, userID)
}
