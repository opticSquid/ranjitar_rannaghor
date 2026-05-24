package journal

import (
	"math"
	"testing"
	"time"
)

func floatEq(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

// --- CalculateTotalCost ---

func TestCalculateTotalCost(t *testing.T) {
	basePrices := map[string]float64{
		"standard":  52.5,
		"special":   120.0,
		"rice":      10.0,
		"roti":      4.0,
		"chicken":   30.0,
		"fish":      20.0,
		"egg":       10.0,
		"vegetable": 15.0,
	}

	tests := []struct {
		name   string
		prices map[string]float64
		req    EntryRequest
		want   float64
	}{
		{
			name:   "standard-no-extras",
			prices: basePrices,
			req:    EntryRequest{HasMainMeal: true, IsSpecial: false},
			want:   52.5,
		},
		{
			name:   "special-with-extras",
			prices: basePrices,
			req: EntryRequest{
				HasMainMeal:       true,
				IsSpecial:         true,
				ExtraRiceQty:      1,
				ExtraRotiQty:      2,
				ExtraChickenQty:   1,
				ExtraFishQty:      0,
				ExtraEggQty:       3,
				ExtraVegetableQty: 1,
			},
			want: 213.0, // 120 + 10 + 8 + 30 + 0 + 30 + 15
		},
		{
			name:   "no-main-with-extras",
			prices: basePrices,
			req: EntryRequest{
				HasMainMeal:  false,
				ExtraRiceQty: 2,
				ExtraRotiQty: 1,
				ExtraFishQty: 1,
			},
			want: 44.0, // 2*10 + 1*4 + 1*20
		},
		{
			name:   "empty-prices",
			prices: map[string]float64{},
			req:    EntryRequest{HasMainMeal: true, ExtraRiceQty: 1, ExtraRotiQty: 1},
			want:   0.0,
		},
		{
			name:   "negative-extras",
			prices: basePrices,
			req:    EntryRequest{HasMainMeal: false, ExtraRiceQty: -1, ExtraRotiQty: 1},
			want:   -6.0, // -1*10 + 1*4
		},
		{
			name:   "large-quantities",
			prices: basePrices,
			req:    EntryRequest{HasMainMeal: true, ExtraEggQty: 1000},
			want:   52.5 + 1000*10.0,
		},
		{
			name:   "special-no-main",
			prices: basePrices,
			req:    EntryRequest{HasMainMeal: false, IsSpecial: true, ExtraRiceQty: 1},
			want:   10.0, // special price ignored when HasMainMeal is false
		},
		{
			name: "fractional-prices",
			prices: map[string]float64{
				"standard": 50.75,
				"roti":     4.25,
			},
			req:  EntryRequest{HasMainMeal: true, ExtraRotiQty: 2},
			want: 59.25, // 50.75 + (2 * 4.25)
		},
		{
			name: "missing-specific-keys",
			prices: map[string]float64{
				"standard": 50.0,
			},
			req:  EntryRequest{HasMainMeal: true, ExtraRiceQty: 2},
			want: 50.0, // missing rice defaults to 0
		},
		{
			name:   "all-zero-extras",
			prices: basePrices,
			req: EntryRequest{
				HasMainMeal:       true,
				ExtraRiceQty:      0,
				ExtraRotiQty:      0,
				ExtraChickenQty:   0,
				ExtraFishQty:      0,
				ExtraEggQty:       0,
				ExtraVegetableQty: 0,
			},
			want: 52.5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateTotalCost(tc.req, tc.prices)
			if !floatEq(got, tc.want) {
				t.Fatalf("%s: want %v, got %v", tc.name, tc.want, got)
			}
		})
	}
}

// --- getCreationTime ---

func TestGetCreationTime_ReturnsSameDateAsInput(t *testing.T) {
	logDate := time.Date(2023, time.October, 5, 0, 0, 0, 0, time.UTC)
	result := getCreationTime(logDate)
	y, m, d := result.Date()
	if y != 2023 || m != time.October || d != 5 {
		t.Fatalf("expected date 2023-10-05, got %v", result)
	}
}

func TestGetCreationTime_PreservesCurrentTimeOfDay(t *testing.T) {
	logDate := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	before := time.Now().UTC()
	result := getCreationTime(logDate)
	after := time.Now().UTC()

	// The time-of-day part should be between before and after (current wall clock)
	resultTOD := result.Hour()*3600 + result.Minute()*60 + result.Second()
	beforeTOD := before.Hour()*3600 + before.Minute()*60 + before.Second()
	afterTOD := after.Hour()*3600 + after.Minute()*60 + after.Second()

	if resultTOD < beforeTOD || resultTOD > afterTOD+1 {
		t.Fatalf("time-of-day %02d:%02d:%02d not within expected window %02d:%02d:%02d – %02d:%02d:%02d",
			result.Hour(), result.Minute(), result.Second(),
			before.Hour(), before.Minute(), before.Second(),
			after.Hour(), after.Minute(), after.Second())
	}
}
