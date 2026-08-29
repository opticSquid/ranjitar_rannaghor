package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (c MealType) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

func (c *MealType) UnmarshalText(text []byte) error {
	val := MealType(text)
	switch val {
	case LUNCH, DINNER:
		*c = val
		return nil
	default:
		return fmt.Errorf("invalid meal type value: %s", string(text))
	}
}

func (c OrderStatus) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

func (c *OrderStatus) UnmarshalText(text []byte) error {
	val := OrderStatus(text)
	switch val {
	case COMPLETED, PENDING, CANCELLED:
		*c = val
		return nil
	default:
		return fmt.Errorf("invalid order status value: %s", string(text))
	}
}

func (c TxnType) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

func (c *TxnType) UnmarshalText(text []byte) error {
	val := TxnType(text)
	switch val {
	case RECHARGE, REFUND, DELIVERY:
		*c = val
		return nil
	default:
		return fmt.Errorf("invalid transaction type value: %s", string(text))
	}
}

func checkUserExist(tx pgx.Tx, ctx context.Context, userId int) (bool, error) {
	var doesExist bool
	err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM PUBLIC.USERS WHERE USER_ID = $1)", userId).Scan(&doesExist)
	if err != nil {
		return false, err
	}
	return doesExist, nil
}

func getUserDetails(tx pgx.Tx, ctx context.Context, userId int) (UserDetails, error) {
	var userDetails UserDetails
	err := tx.QueryRow(ctx, `SELECT NAME, MOBILE_NO FROM PUBLIC.USERS WHERE USER_ID = $1`, userId).Scan(&userDetails.Name, &userDetails.MobileNo)
	if err != nil {
		return UserDetails{}, err
	}
	return userDetails, nil
}

func fetchBalance(tx pgx.Tx, ctx context.Context, userId int, endTs time.Time) (float64, error) {
	var balance float64
	err := tx.QueryRow(ctx, `SELECT
    COALESCE(SUM(
        CASE txn_type
            WHEN 'recharge' THEN amount
            WHEN 'refund'   THEN amount
            WHEN 'delivery' THEN -amount
        END
    ), 0) AS balance_at_ts
	FROM wallet_transactions
	WHERE user_id = $1
  AND txn_timestamp <= $2`, userId, endTs).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func fetchOrders(tx pgx.Tx, ctx context.Context, startDate time.Time, endDate time.Time) ([]Delivery, error) {
	rows, err := tx.Query(ctx, `SELECT `)
}
