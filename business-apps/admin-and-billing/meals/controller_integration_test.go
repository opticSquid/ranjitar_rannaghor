package meals

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testdb.Setup()
	defer testdb.Teardown()
	m.Run()
}

func TestCreateAndGetMeals(t *testing.T) {
	testdb.ResetData()

	m := MealPrice{ItemName: "TestMeal", Price: 9.99}
	body, _ := json.Marshal(m)
	req := httptest.NewRequest("POST", "/meals", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	CreateMeal(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// fetch meals
	req2 := httptest.NewRequest("GET", "/meals", nil)
	rr2 := httptest.NewRecorder()
	GetMeals(rr2, req2)
	require.Equal(t, http.StatusOK, rr2.Code)

	var meals []MealPrice
	json.NewDecoder(rr2.Body).Decode(&meals)
	require.True(t, len(meals) >= 1)
}

func TestCreatePriceAndFetchHistory(t *testing.T) {
	testdb.ResetData()
	// use seeded 'standard' item
	reqBody := map[string]interface{}{
		"price":          65.5,
		"effective_from": time.Now().UTC().Add(2 * time.Minute).Format(time.RFC3339),
		"created_by":     "test",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/meals/standard/prices", bytes.NewBuffer(b))
	rr := httptest.NewRecorder()

	CreatePrice(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	// fetch history
	req2 := httptest.NewRequest("GET", "/meals/standard/prices", nil)
	rr2 := httptest.NewRecorder()
	GetPriceHistory(rr2, req2)
	require.Equal(t, http.StatusOK, rr2.Code)
	var entries []PriceHistoryEntry
	json.NewDecoder(rr2.Body).Decode(&entries)
	require.True(t, len(entries) >= 1)
}
