package expenses

import "time"

type Expense struct {
	ExpenseID   int       `json:"expense_id"`
	ExpenseDate time.Time `json:"expense_date"`
	Reason      string    `json:"reason"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}
