package meals

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
