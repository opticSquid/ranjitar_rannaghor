package billing

import (
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/journal"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/users"
)

type BillReport struct {
	User           users.User         `json:"user"`
	StartDate      time.Time          `json:"start_date"`
	EndDate        time.Time          `json:"end_date"`
	Logs           []journal.DailyLog `json:"logs"`
	TotalSpent     float64            `json:"total_spent"`
	TotalRecharges float64            `json:"total_recharges"`
	OpeningBalance float64            `json:"opening_balance"`
	ClosingBalance float64            `json:"closing_balance"`
}
