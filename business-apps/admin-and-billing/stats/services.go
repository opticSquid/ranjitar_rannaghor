package stats

import (
	"context"
	"time"
)

func GetDashboardStatsService(ctx context.Context) (DashboardStats, error) {
	var stats DashboardStats
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	err := FetchDashboardStatsFromDB(ctx, &stats, firstOfMonth)
	if err != nil {
		return stats, err
	}

	stats.NetProfit = stats.TotalRevenue - stats.TotalExpenses
	return stats, nil
}

func GetAnalyticsStatsService(ctx context.Context) (AnalyticsStats, error) {
	var stats AnalyticsStats
	now := time.Now()
	startDate := now.AddDate(0, 0, -30)

	err := FetchAnalyticsStatsFromDB(ctx, &stats, startDate)
	if err != nil {
		return stats, err
	}

	if stats.TotalRevenue > 0 {
		stats.ProfitPercentage = ((stats.TotalRevenue - stats.TotalExpenses) / stats.TotalRevenue) * 100
	}
	return stats, nil
}
