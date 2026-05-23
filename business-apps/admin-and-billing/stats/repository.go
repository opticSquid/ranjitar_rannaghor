package stats

import (
	"context"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func FetchDashboardStatsFromDB(ctx context.Context, stats *DashboardStats, firstOfMonth time.Time) error {
	dbPool := database.GetDbConn()

	// 1. Total & Monthly Revenue
	err := dbPool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(TOTAL_COST), 0),
			COALESCE(SUM(CASE WHEN LOG_DATE >= $1 THEN TOTAL_COST ELSE 0 END), 0)
		FROM DAILY_LOGS
	`, firstOfMonth).Scan(&stats.TotalRevenue, &stats.MonthlyRevenue)
	if err != nil {
		return err
	}

	// 2. Total & Monthly Expenses
	err = dbPool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(AMOUNT), 0),
			COALESCE(SUM(CASE WHEN EXPENSE_DATE >= $1 THEN AMOUNT ELSE 0 END), 0)
		FROM EXPENSES
	`, firstOfMonth).Scan(&stats.TotalExpenses, &stats.MonthlyExpenses)
	if err != nil {
		return err
	}

	// 4. Active Customers Count
	err = dbPool.QueryRow(ctx, `SELECT COUNT(*) FROM USERS`).Scan(&stats.ActiveCustomers)
	if err != nil {
		return err
	}

	// 5. Wallet Pool
	err = dbPool.QueryRow(ctx, `
		SELECT COALESCE(SUM(balance), 0) FROM (
			SELECT DISTINCT ON (user_id) balance_after AS balance
			FROM WALLET_TRANSACTIONS
			WHERE STATUS = 'confirmed'
			ORDER BY user_id, created_at DESC, txn_id DESC
		) sub
	`).Scan(&stats.WalletPool)
	return err
}

func FetchAnalyticsStatsFromDB(ctx context.Context, stats *AnalyticsStats, startDate time.Time) error {
	dbPool := database.GetDbConn()

	// 1. Revenue Trends
	revenueRows, err := dbPool.Query(ctx, `
		SELECT LOG_DATE, SUM(TOTAL_COST)
		FROM DAILY_LOGS
		WHERE LOG_DATE >= $1
		GROUP BY LOG_DATE
		ORDER BY LOG_DATE ASC
	`, startDate)
	if err != nil {
		return err
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

	// 2. Expense Trends
	expenseRows, err := dbPool.Query(ctx, `
		SELECT EXPENSE_DATE, SUM(AMOUNT)
		FROM EXPENSES
		WHERE EXPENSE_DATE >= $1
		GROUP BY EXPENSE_DATE
		ORDER BY EXPENSE_DATE ASC
	`, startDate)
	if err != nil {
		return err
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

	// Fill in Trends
	for i := 0; i <= 30; i++ {
		d := startDate.AddDate(0, 0, i).Format("2006-01-02")
		stats.Trends = append(stats.Trends, TrendPoint{
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
		return err
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
		return err
	}
	stats.Shifts["Lunch"] = lunchCount
	stats.Shifts["Dinner"] = dinnerCount

	return nil
}
