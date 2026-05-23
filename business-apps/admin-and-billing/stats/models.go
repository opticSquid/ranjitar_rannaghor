package stats

type DashboardStats struct {
	TotalRevenue    float64 `json:"total_revenue"`
	TotalExpenses   float64 `json:"total_expenses"`
	NetProfit       float64 `json:"net_profit"`
	MonthlyRevenue  float64 `json:"monthly_revenue"`
	MonthlyExpenses float64 `json:"monthly_expenses"`
	ActiveCustomers int     `json:"active_customers"`
	WalletPool      float64 `json:"wallet_pool"`
}

type TrendPoint struct {
	Date     string  `json:"date"`
	Revenue  float64 `json:"revenue"`
	Expenses float64 `json:"expenses"`
}

type AnalyticsStats struct {
	Trends           []TrendPoint   `json:"trends"`
	MealTypes        map[string]int `json:"meal_types"`
	Shifts           map[string]int `json:"shifts"`
	TotalRevenue     float64        `json:"total_revenue"`
	TotalExpenses    float64        `json:"total_expenses"`
	ProfitPercentage float64        `json:"profit_percentage"`
}
