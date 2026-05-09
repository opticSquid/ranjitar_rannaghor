package stats

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/soumalya/food-delivery-admin/database"
	"github.com/soumalya/food-delivery-admin/model"
)

func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	var stats model.DashboardStats
	ctx := r.Context()

	// 1. Total & Monthly Revenue (from DAILY_LOGS)
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	dbPool := database.GetDbConn()
	err := dbPool.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(TOTAL_COST), 0),
			COALESCE(SUM(CASE WHEN LOG_DATE >= $1 THEN TOTAL_COST ELSE 0 END), 0)
		FROM DAILY_LOGS
	`, firstOfMonth).Scan(&stats.TotalRevenue, &stats.MonthlyRevenue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Total & Monthly Expenses
	err = dbPool.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(AMOUNT), 0),
			COALESCE(SUM(CASE WHEN EXPENSE_DATE >= $1 THEN AMOUNT ELSE 0 END), 0)
		FROM EXPENSES
	`, firstOfMonth).Scan(&stats.TotalExpenses, &stats.MonthlyExpenses)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Profit
	stats.NetProfit = stats.TotalRevenue - stats.TotalExpenses

	// 4. Active Customers Count
	err = dbPool.QueryRow(ctx, `SELECT COUNT(*) FROM USERS`).Scan(&stats.ActiveCustomers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Wallet Pool (from WALLET_TRANSACTIONS — WALLET table is deprecated)
	err = dbPool.QueryRow(ctx, `
		SELECT COALESCE(SUM(balance), 0) FROM (
			SELECT DISTINCT ON (user_id) balance_after AS balance
			FROM WALLET_TRANSACTIONS
			WHERE STATUS = 'confirmed'
			ORDER BY user_id, created_at DESC, txn_id DESC
		) sub
	`).Scan(&stats.WalletPool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func GetAnalyticsStats(w http.ResponseWriter, r *http.Request) {
	var stats model.AnalyticsStats
	ctx := r.Context()
	now := time.Now()
	startDate := now.AddDate(0, 0, -30)

	// 1. Revenue Trends (last 30 days)
	dbPool := database.GetDbConn()
	revenueRows, err := dbPool.Query(ctx, `
		SELECT LOG_DATE, SUM(TOTAL_COST) 
		FROM DAILY_LOGS 
		WHERE LOG_DATE >= $1 
		GROUP BY LOG_DATE 
		ORDER BY LOG_DATE ASC
	`, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer revenueRows.Close()

	revenueMap := make(map[string]float64)
	for revenueRows.Next() {
		var date time.Time
		var amount float64
		revenueRows.Scan(&date, &amount)
		revenueMap[date.Format("2006-01-02")] = amount
		stats.TotalRevenue += amount
	}

	// 2. Expense Trends (last 30 days)
	expenseRows, err := dbPool.Query(ctx, `
		SELECT EXPENSE_DATE, SUM(AMOUNT) 
		FROM EXPENSES 
		WHERE EXPENSE_DATE >= $1 
		GROUP BY EXPENSE_DATE 
		ORDER BY EXPENSE_DATE ASC
	`, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer expenseRows.Close()

	expenseMap := make(map[string]float64)
	for expenseRows.Next() {
		var date time.Time
		var amount float64
		expenseRows.Scan(&date, &amount)
		expenseMap[date.Format("2006-01-02")] = amount
		stats.TotalExpenses += amount
	}

	// Fill in Trends for each of the last 30 days
	for i := 0; i <= 30; i++ {
		d := startDate.AddDate(0, 0, i).Format("2006-01-02")
		stats.Trends = append(stats.Trends, model.TrendPoint{
			Date:     d,
			Revenue:  revenueMap[d],
			Expenses: expenseMap[d],
		})
	}

	// 3. Meal Type Distribution
	stats.MealTypes = make(map[string]int)
	var standardCount, specialCount int
	err = dbPool.QueryRow(ctx, `
		SELECT 
			COUNT(CASE WHEN IS_SPECIAL = false THEN 1 END),
			COUNT(CASE WHEN IS_SPECIAL = true THEN 1 END)
		FROM DAILY_LOGS
	`).Scan(&standardCount, &specialCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stats.MealTypes["Standard"] = standardCount
	stats.MealTypes["Special"] = specialCount

	// 4. Shift Distribution
	stats.Shifts = make(map[string]int)
	var lunchCount, dinnerCount int
	err = dbPool.QueryRow(ctx, `
		SELECT 
			COUNT(CASE WHEN MEAL_TYPE = 'lunch' THEN 1 END),
			COUNT(CASE WHEN MEAL_TYPE = 'dinner' THEN 1 END)
		FROM DAILY_LOGS
	`).Scan(&lunchCount, &dinnerCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stats.Shifts["Lunch"] = lunchCount
	stats.Shifts["Dinner"] = dinnerCount

	// 5. Profit Percentage
	if stats.TotalRevenue > 0 {
		stats.ProfitPercentage = ((stats.TotalRevenue - stats.TotalExpenses) / stats.TotalRevenue) * 100
	}

	json.NewEncoder(w).Encode(stats)
}
