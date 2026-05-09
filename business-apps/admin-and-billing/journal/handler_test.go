package journal

import (
	"math"
	"testing"

	"github.com/soumalya/food-delivery-admin/model"
)

func floatEq(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

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
		req    model.EntryRequest
		want   float64
	}{
		{
			name:   "standard-no-extras",
			prices: basePrices,
			req:    model.EntryRequest{HasMainMeal: true, IsSpecial: false},
			want:   52.5,
		},
		{
			name:   "special-with-extras",
			prices: basePrices,
			req: model.EntryRequest{
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
			req: model.EntryRequest{
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
			req:    model.EntryRequest{HasMainMeal: true, ExtraRiceQty: 1, ExtraRotiQty: 1},
			want:   0.0, // missing price keys default to zero
		},
		{
			name:   "negative-extras",
			prices: basePrices,
			req:    model.EntryRequest{HasMainMeal: false, ExtraRiceQty: -1, ExtraRotiQty: 1},
			want:   -6.0, // -1*10 + 1*4
		},
		{
			name:   "large-quantities",
			prices: basePrices,
			req:    model.EntryRequest{HasMainMeal: true, ExtraEggQty: 1000},
			want:   52.5 + 1000*10.0,
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
